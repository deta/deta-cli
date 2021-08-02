package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
	runtimeManager, err := runtime.NewManager(nil, false)
	if err != nil {
		return "", err
	}

	u, err := getUserInfo(runtimeManager, client)
	if err != nil {
		return "", err
	}

	res, err := client.GetProjects(&api.GetProjectsRequest{
		SpaceID: u.DefaultSpace,
	})

	for _, pr := range res.Projects {
		if pr.ID == p.Project {
			p.Project = pr.Name
		}
	}

	if err != nil {
		return "", err
	}

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

func parseDateTime(str string) (int64, error) {
	str = strings.Trim(str, SPACE)
	if len(str) != 0 {
		dateTime, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return 0, fmt.Errorf("invalid date time format(RFC3339) %s", str)
		}

		return dateTime.UnixNano() / int64(time.Millisecond), nil
	}

	return 0, nil
}
