/*
  Onix Config Manager - Artisan
  Copyright (c) 2018-2021 by www.gatblau.org
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package deploy

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
)

// AppManifest the application manifest that is made up of one or more service manifests
type AppManifest struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Version     string   `yaml:"version"`
	Services    []SvcRef `yaml:"services"`
	Var         Vars     `yaml:"var,omitempty"`
}

type SvcRef struct {
	// the name of the service
	Name string `yaml:"name,omitempty"`
	// the service description
	Description string `yaml:"description"`
	// the uri of the service manifest
	URI string `yaml:"uri,omitempty"`
	// the uri of the database schema definition (if any)
	SchemaURI string `yaml:"schema_uri,omitempty"`
	// the URI of the service image containing the service manifest
	Image string `yaml:"image,omitempty"`
	// whether this service should not be publicly exposed, by default is false
	Private bool `yaml:"private,omitempty"`
	// the service port, if not specified, the application port (in the service manifest) is used
	Port string `yaml:"port,omitempty"`
	// the service manifest loaded from remote image
	Info *SvcManifest `yaml:"service,omitempty"`
	// the other services it depends on
	DependsOn []string `yaml:"depends_on,omitempty"`
}

// NewAppMan creates a new application manifest from an URI (supported schemes are http(s):// and file://
func NewAppMan(uri string) (*AppManifest, error) {
	if ok, path := isFile(uri); ok {
		return loadFromFile(path)
	}
	if isURL(uri) {
		return loadFromURL(uri)
	}
	return nil, fmt.Errorf("invalid URI value '%s': should start with either file://, http:// or https://\n", uri)
}

// Explode augments an app manifest that has remote references to service manifests
func (m *AppManifest) Explode() (*AppManifest, error) {
	var err error
	// create a copy of the passed in light manifest to become the exploded version
	appMan := new(AppManifest)
	_ = m.deepCopy(appMan)
	// validate the app manifest
	if err = m.validate(); err != nil {
		return nil, err
	}
	// loop through
	var svcMan *SvcManifest
	for i, svc := range m.Services {
		// image only
		if len(svc.Image) > 0 && len(svc.URI) == 0 {
			svcMan, err = loadSvcManFromImage(svc)
			if err != nil {
				return nil, fmt.Errorf("cannot load service manifest for '%s': %s\n", svc.Image, err)
			}
		} else if len(svc.Image) > 0 && len(svc.URI) > 0 {
			svcMan, err = loadSvcManFromURI(svc)
			if err != nil {
				return nil, fmt.Errorf("cannot load service manifest for '%s': %s\n", svc.Image, err)
			}
		}
		appMan.Services[i].Info = svcMan
		appMan.Services[i].Name = svcMan.Name
	}
	return appMan, nil
}

func (m *AppManifest) Wire() (*AppManifest, error) {
	appMan := new(AppManifest)
	_ = m.deepCopy(appMan)
	for six, service := range m.Services {
		for vix, v := range service.Info.Var {
			// if the variable is a function expression
			if strings.HasPrefix(strings.Replace(v.Value, " ", "", -1), "{{fx=") {
				content := v.Value[len("{{fx=") : len(v.Value)-2]
				parts := strings.Split(content, ":")
				// qualifies the name of the variable with the service name
				vName := fmt.Sprintf("${%s_%s}", strings.ToUpper(strings.Replace(service.Name, "-", "_", -1)), v.Name)
				switch strings.ToLower(parts[0]) {
				case "pwd":
					subParts := strings.Split(parts[1], ",")
					length, _ := strconv.Atoi(subParts[0])
					symbols, _ := strconv.ParseBool(subParts[1])
					appMan.Services[six].Info.Var[vix].Value = vName
					appMan.Var.Items = append(appMan.Var.Items, AppVar{
						Name:        vName,
						Description: v.Description,
						Value:       RandomPwd(length, symbols),
						Secret:      true,
						Service:     strings.ToUpper(service.Name),
					})
				case "name":
					number, _ := strconv.Atoi(parts[1])
					appMan.Services[six].Info.Var[vix].Value = vName
					appMan.Var.Items = append(appMan.Var.Items, AppVar{
						Name:        vName,
						Description: v.Description,
						Value:       RandomName(number),
						Secret:      false,
						Service:     strings.ToUpper(service.Name),
					})
				default:
					return nil, fmt.Errorf("invalid function %s='%s' in service '%s'\n", v.Name, v.Value, service.Name)
				}
			} else // if the variable is a binding
			if strings.HasPrefix(strings.Replace(v.Value, " ", "", -1), "{{bind=") {
				content := v.Value[len("{{bind=") : len(v.Value)-2]
				parts := strings.Split(content, ":")
				switch len(parts) {
				case 1:
					svcName := parts[0]
					// check the name exists
					found := false
					for _, s := range m.Services {
						if s.Name == svcName {
							found = true
							break
						}
					}
					if !found {
						return nil, fmt.Errorf("invalid service name '%s' => %s='%s' in service '%s'\n", svcName, v.Name, v.Value, service.Name)
					}
					appMan.Services[six].Info.Var[vix].Value = svcName
				case 2:
					switch parts[1] {
					case "schema_uri":
						if uri := m.getSchemaURI(parts[0]); len(uri) > 0 {
							appMan.Services[six].Info.Var[vix].Value = uri
						} else {
							return nil, fmt.Errorf("variable %s='%s' in service '%s' request schema_ui from service '%s' but is missing\n", v.Name, v.Value, service.Name, parts[0])
						}
					}
				default:
					return nil, fmt.Errorf("invalid binding %s='%s' in service '%s'\n", v.Name, v.Value, service.Name)
				case 3:
					switch parts[1] {
					case "var":
						if m.varExists(parts[2]) {
							appMan.Services[six].Info.Var[vix].Value = strings.ToUpper(fmt.Sprintf("${%s_%s}", parts[0], parts[2]))
							appMan.Services[six].DependsOn = addDependency(appMan.Services[six].DependsOn, parts[0])
						} else {
							return nil, fmt.Errorf("cannot find variable %s='%s' in service '%s'\n", v.Name, v.Value, service.Name)
						}
					}
				}
			}
		}
	}
	return appMan, nil
}

func addDependency(dependsOn []string, svc string) []string {
	result := make([]string, len(dependsOn))
	copy(result, dependsOn)
	exists := false
	for _, d := range result {
		if d == svc {
			exists = true
			break
		}
	}
	if !exists {
		result = append(result, svc)
	}
	return result
}

func (m *AppManifest) getSchemaURI(svc string) string {
	for _, service := range m.Services {
		if service.Name == svc && len(service.SchemaURI) > 0 {
			return service.SchemaURI
		}
	}
	return ""
}

func (m *AppManifest) varExists(varName string) bool {
	for _, service := range m.Services {
		for _, v := range service.Info.Var {
			if v.Name == varName {
				return true
			}
		}
	}
	return false
}

func (m *AppManifest) validate() error {
	for _, svc := range m.Services {
		// case of manifest embedded in docker image then no URI is needed (image only)
		// case of manifest in git repo (uri + image required)
		// so cases to avoid is uri only
		if len(svc.Image) == 0 && len(svc.URI) > 0 {
			return fmt.Errorf("invalid entry for service '%s' manifest in application manifest: only one of Image or URI attributes must be specified\n", svc.Name)
		}
		// or uri & image not provided
		if len(svc.Image) == 0 && len(svc.URI) == 0 {
			return fmt.Errorf("invalid entry for service '%s' manifest in application manifest: either one of Image or URI attributes must be specified\n", svc.Name)
		}
	}
	return nil
}

func (m *AppManifest) deepCopy(dst interface{}) error {
	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(m); err != nil {
		return err
	}
	return gob.NewDecoder(&buffer).Decode(dst)
}
