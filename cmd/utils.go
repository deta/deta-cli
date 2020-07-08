package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/deta/deta-cli/runtime"
)

var (
	// set with make file during compilation
	gatewayDomain string
)

type progDetailsOutput struct {
	Name     string   `json:"name"`
	Runtime  string   `json:"runtime"`
	Endpoint string   `json:"endpoint"`
	Deps     []string `json:"dependencies,omitempty"`
	Envs     []string `json:"environment_variables,omitempty"`
	Visor    string   `json:"visor"`
	Auth     string   `json:"http_auth"`
	Cron     string   `json:"cron,omitempty"`
}

func progInfoToOutput(p *runtime.ProgInfo) (string, error) {
	o := progDetailsOutput{
		Name:    p.Name,
		Runtime: p.Runtime,
		Deps:    p.Deps,
		Envs:    p.Envs,
		Visor:   "enabled",
		Auth:    "enabled",
		Cron:    p.Cron,
	}
	o.Endpoint = fmt.Sprintf("https://%s.%s", p.Path, gatewayDomain)
	if p.Visor == "off" {
		o.Visor = "disabled"
	}
	if p.Public {
		o.Auth = "disabled"
	}

	po, err := prettyPrint(o)
	if err != nil {
		return "", err
	}
	return po, nil
}

func prettyPrint(data interface{}) (string, error) {
	marshalled, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return "", err
	}
	return string(marshalled), nil
}

func removeFromSlice(slice []string, toRemove string) []string {
	for i, s := range slice {
		if s == toRemove {
			slice[i] = slice[len(slice)-1]
			return slice[:len(slice)-1]
		}
	}
	return slice
}

func inSlice(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
