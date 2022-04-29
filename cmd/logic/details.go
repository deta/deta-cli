package logic

import (
	"fmt"
	"os"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
)

func Details(client *api.DetaClient, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if len(args) != 0 {
		wd = args[0]
	}
	runtimeManager, err := runtime.NewManager(&wd, false)
	if err != nil {
		return err
	}
	isInitialzied, err := runtimeManager.IsInitialized()
	if err != nil {
		return err
	}
	if !isInitialzied {
		return fmt.Errorf("no deta micro initialized in '%s'", wd)
	}
	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}
	if progInfo == nil {
		return fmt.Errorf("failed to get deta micro details")
	}
	u, err := GetUserInfo(runtimeManager, client)
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
		if p.ID == progInfo.Project {
			progInfo.Project = p.Name
		}
	}
	output, err := progInfoToOutput(progInfo)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
