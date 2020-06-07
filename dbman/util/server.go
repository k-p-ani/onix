//   Onix Config DatabaseProvider - Dbman
//   Copyright (c) 2018-2020 by www.gatblau.org
//   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
//   Contributors to this project, hereby assign copyright in this code to the project,
//   to be licensed under the same terms as the rest of the code.
package util

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

type Server struct {
	cfg   *AppCfg
	start time.Time
}

func NewServer(cfg *AppCfg) *Server {
	return &Server{
		cfg: cfg,
	}
}

func (s *Server) Serve() {
	// compute the time the server is called
	s.start = time.Now()

	mux := mux.NewRouter()
	mux.Use(s.loggingMiddleware)
	mux.Use(s.authenticationMiddleware)

	// registers web handlers
	fmt.Printf("? I am registering http handlers\n")
	mux.HandleFunc("/", s.liveHandler).Methods("GET")
	mux.HandleFunc("/ready", s.readyHandler).Methods("GET")
	mux.HandleFunc("/db/init", s.initHandler).Methods("POST")
	mux.HandleFunc("/db/deploy/{appVersion}", s.deployHandler).Methods("POST")

	if s.cfg.GetBool(HttpMetrics) {
		// prometheus metrics
		fmt.Printf("? I am registering the metrics publication handler '/metrics'\n")
		mux.Handle("/metrics", promhttp.Handler()).Methods("GET")
	}

	// starts the server
	s.listen(mux)
}

// a liveliness probe to prove the http service is listening
func (s *Server) liveHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK"))
	if err != nil {
		fmt.Printf("!!! I cannot write response: %v", err)
	}
}

// a readyness probe to prove DbMan is ready to accept calls
func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	ready, err := DM.CheckReady()
	if !ready {
		fmt.Printf("! I am not ready: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
		}
	} else {
		_, _ = w.Write([]byte("OK"))
	}
}

func (s *Server) initHandler(w http.ResponseWriter, r *http.Request) {
	// deploy the schema and functions
	err := DM.InitialiseDb()
	// return an error if failed
	if err != nil {
		s.writeError(w, err)
	} else {
		_, err = w.Write([]byte(fmt.Sprintf("? Initialisation complete")))
		if err != nil {
			fmt.Printf("!!! I failed to write error to response: %v", err)
		}
	}
}

// deploy a schema
func (s *Server) deployHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appVersion := vars["appVersion"]
	// deploy the schema and functions
	err := DM.Deploy(appVersion)
	// return an error if failed
	if err != nil {
		s.writeError(w, err)
	} else {
		_, err = w.Write([]byte(fmt.Sprintf("? Deployment complete")))
		if err != nil {
			fmt.Printf("!!! I failed to write error to response: %v", err)
		}
	}
}

// creates a new Basic Authentication Token
func (s *Server) newBasicToken(user string, pwd string) string {
	return fmt.Sprintf(
		"Basic %s",
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, pwd))))
}

func (s *Server) listen(handler http.Handler) {
	// creates an http server listening on the specified TCP port
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", s.cfg.Get(HttpPort)),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  time.Second * 60,
		Handler:      handler,
	}

	// runs the server asynchronously
	go func() {
		fmt.Printf("? I am listening on :%s\n", s.cfg.Get(HttpPort))
		fmt.Printf("? I have taken %v to start\n", time.Since(s.start))
		if err := server.ListenAndServe(); err != nil {
			fmt.Printf("! Stopping the server: %v\n", err)
		}
	}()

	// creates a channel to pass a SIGINT (ctrl+C) kernel signal with buffer capacity 1
	stop := make(chan os.Signal, 1)

	// sends any SIGINT signal to the stop channel
	signal.Notify(stop, os.Interrupt)

	// waits for the SIGINT signal to be raised (pkill -2)
	<-stop

	// gets a context with some delay to shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// releases resources if main completes before the delay period elapses
	defer cancel()

	// on error shutdown
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("? I am shutting down due to an error: %v\n", err)
	}
}

func (s *Server) writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	_, err = w.Write([]byte(err.Error()))
	if err != nil {
		fmt.Printf("!!! I failed to write error to response: %v", err)
	}
}

// middleware
// log http requests to stdout
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("? I received http request: %s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

// determines if the request is authenticated
func (s *Server) authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// gets the authentication mode
		authMode := s.cfg.Get(HttpAuthMode)

		// if there is a username and password
		if len(authMode) > 0 && strings.ToLower(authMode) == "basic" {
			if r.Header.Get("Authorization") == "" {
				// if no authorisation header is passed, then it prompts a client browser to authenticate
				w.Header().Set("WWW-Authenticate", `Basic realm="dbman"`)
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Printf("? I have received an unauthorised request from: '%v'\n", r.RemoteAddr)
			} else {
				// authenticate the request
				requiredToken := s.newBasicToken(s.cfg.Get(HttpUsername), s.cfg.Get(HttpPassword))
				providedToken := r.Header.Get("Authorization")
				// if the authentication fails
				if providedToken != requiredToken {
					// Write an error and stop the handler chain
					http.Error(w, "Forbidden", http.StatusForbidden)
				}
			}
		}
		// Pass down the request to the next middleware (or final handler)
		next.ServeHTTP(w, r)
	})
}