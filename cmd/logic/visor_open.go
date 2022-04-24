package logic

import (
	"fmt"
	"os"
	"os/exec"
	rt "runtime"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
)

var (
	// set with Makefile during compilation
	visorURL string
)

func VisorOpen(client *api.DetaClient, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if len(args) > 0 {
		wd = args[0]
	}
	runtimeManager, err := runtime.NewManager(&wd, false)
	if err != nil {
		return err
	}

	isInitialized, err := runtimeManager.IsInitialized()
	if err != nil {
		return err
	}

	if !isInitialized {
		return fmt.Errorf(fmt.Sprintf("no deta micro present in '%s'", wd))
	}

	userInfo, err := getUserInfo(runtimeManager, client)
	if err != nil {
		return err
	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}

	resp, err := client.GetProjects(&api.GetProjectsRequest{
		SpaceID: userInfo.DefaultSpace,
	})
	if err != nil {
		return err
	}
	var progProject string
	for _, p := range resp.Projects {
		if p.ID == progInfo.Project {
			progProject = p.Name
		}
	}

	visorEndpoint := fmt.Sprintf("%s/?space=%s&project=%s&micro=%s",
		visorURL,
		userInfo.DefaultSpaceName,
		progProject,
		progInfo.Name,
	)
	fmt.Println("Opening visor in the browser...")
	return openVisorPage(visorEndpoint)
}

func openVisorPage(url string) error {
	switch rt.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}
