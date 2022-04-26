package logic

import (
	"fmt"
	"strings"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
)

func Pull(client *api.DetaClient, forcePull bool, args []string) error {
	runtimeManager, err := runtime.NewManager(nil, false)
	if err != nil {
		return err
	}

	isInitialized, err := runtimeManager.IsInitialized()
	if err != nil {
		return err
	}
	if !isInitialized {
		return fmt.Errorf("no deta micro initialized in current directory")
	}

	if !forcePull {
		fmt.Println("Files already present may be overwritten. Continue? [y/n]")
		var cont string
		fmt.Scanf("%s", &cont)
		if strings.ToLower(cont) != "y" {
			fmt.Println("Pull aborted")
			return nil
		}

	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}

	o, err := client.DownloadProgram(&api.DownloadProgramRequest{
		ProgramID: progInfo.ID,
		Runtime:   progInfo.Runtime,
		Account:   progInfo.Account,
		Region:    progInfo.Region,
	})

	if err != nil {
		return err
	}

	err = runtimeManager.WriteProgramFiles(o.ZipFile, nil, true, progInfo.Runtime)
	if err != nil {
		return err
	}
	err = runtimeManager.StoreProgInfo(progInfo)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully pulled latest deployed code")
	return nil
}
