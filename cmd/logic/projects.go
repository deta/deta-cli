package logic

import (
	"fmt"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
)

func ListProjects(client *api.DetaClient, args []string) error {
	runtimeManager, err := runtime.NewManager(nil, false)
	if err != nil {
		return err
	}

	u, err := getUserInfo(runtimeManager, client)
	if err != nil {
		return err
	}

	res, err := client.GetProjects(&api.GetProjectsRequest{
		SpaceID: u.DefaultSpace,
	})

	if err != nil {
		return err
	}

	for _, p := range res.Projects {
		p.ID = ""
	}

	output, err := prettyPrint(res.Projects)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
