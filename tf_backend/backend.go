/*
   Onix Config Manager - OxTerra - Terraform Http Backend for Onix
   Copyright (c) 2018-2020 by www.gatblau.org

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
   Unless required by applicable law or agreed to in writing, software distributed under
   the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
   either express or implied.
   See the License for the specific language governing permissions and limitations under the License.

   Contributors to this project, hereby assign copyright in this code to the project,
   to be licensed under the same terms as the rest of the code.
*/
package main

import (
	"errors"
	"fmt"
	. "github.com/gatblau/oxc"
	"github.com/rs/zerolog"
)

type Backend struct {
	config *Config
	client *Client
	ready  bool
}

// start the backend process
func (backend *Backend) start() error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	var err error

	// load the configuration file
	if c, err := NewConfig(); err == nil {
		backend.config = c
	} else {
		return err
	}

	// initialises the Onix REST client
	backend.client = &Client{BaseURL: backend.config.Ox.URL}

	cfg := backend.config.Ox
	switch cfg.AuthMode {
	case "none":
		// ensure no token value is specified
		backend.client.SetAuthToken("")

	case "basic":
		// sets a basic authentication token
		backend.client.SetAuthToken(backend.client.NewBasicToken(cfg.Username, cfg.Password))

	case "oidc":
		// sets an OAuth Bearer token
		bearerToken, err := backend.client.GetBearerToken(cfg.TokenURI, cfg.ClientId, cfg.ClientSecret, cfg.Username, cfg.Password)
		if err != nil {
			return err
		}
		backend.client.SetAuthToken(bearerToken)

	default:
		// can't recognise the auth_mode provided
		return errors.New(fmt.Sprintf("auth_mode = '%s' is not valid value. Use either 'none', 'basic' or 'oidc'.", cfg.AuthMode))
	}

	// checks if a meta model for Terraform is defined in Onix
	model := NewTerraModel(backend.client)
	err = model.create()
	if err != nil {
		return err
	}

	// the backend is now ready to receive http connections
	backend.ready = true

	// start the service listener
	svc := NewTerraService(*backend)
	svc.Start()
	return nil
}
