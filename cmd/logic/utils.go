package logic

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
)

var (
	// set with make file during compilation
	gatewayDomain string
)

type progDetailsOutput struct {
	Name     string   `json:"name"`
	ID       string   `json:"id"`
	Project  string   `json:"project"`
	Runtime  string   `json:"runtime"`
	Endpoint string   `json:"endpoint"`
	Region   string   `json:"region"`
	Deps     []string `json:"dependencies,omitempty"`
	Envs     []string `json:"environment_variables,omitempty"`
	Visor    string   `json:"visor"`
	Auth     string   `json:"http_auth"`
	Cron     string   `json:"cron,omitempty"`
}

func progInfoToOutput(p *runtime.ProgInfo) (string, error) {
	o := progDetailsOutput{
		Name:    p.Name,
		ID:      p.ID,
		Project: p.Project,
		Runtime: p.Runtime,
		Region:  p.Region,
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

func areSlicesEqualNoOrder(a, b []string) bool {
	// fail fast if len not equal before sorting
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	return reflect.DeepEqual(a, b)
}

// get user info from local storage if cached otherwise from server
// saves user info to local storage if not cached
func getUserInfo(rm *runtime.Manager, client *api.DetaClient) (*runtime.UserInfo, error) {
	u, err := rm.GetUserInfo()
	if err != nil {
		return nil, err
	}
	if u != nil {
		return u, nil
	}

	// fall back to server
	userInfo, err := client.GetUserInfo()
	if err != nil {
		return nil, err
	}

	// save user info
	u = &runtime.UserInfo{
		DefaultSpace:     userInfo.DefaultSpace,
		DefaultSpaceName: userInfo.DefaultSpaceName,
		DefaultProject:   userInfo.DefaultProject,
	}
	go rm.StoreUserInfo(u)
	return u, nil
}

// parseRuntime takes runtimeName as string and returns Runtine struct
func parseRuntime(runtimeName string) (*runtime.Runtime, error) {
	newRuntimeName := runtimeName
	if strings.Contains(runtimeName, runtime.Node) {
		newRuntimeName = fmt.Sprintf("%s.x", runtimeName)
	}

	progRuntime, err := runtime.CheckRuntime(newRuntimeName)
	if err != nil {
		return nil, fmt.Errorf("%s '%s'", err.Error(), runtimeName)
	}

	return progRuntime, nil
}
