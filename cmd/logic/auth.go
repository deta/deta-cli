package logic

import (
	"fmt"
	"os"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
)

func UpdateAuth(client *api.DetaClient, value bool, args []string) error {
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

	isInitialized, err := runtimeManager.IsInitialized()
	if err != nil {
		return err
	}

	if !isInitialized {
		return fmt.Errorf("no deta micro initialized in '%s'", wd)
	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}

	err = client.UpdateAuth(&api.UpdateAuthRequest{
		ProgramID: progInfo.ID,
		AuthValue: value,
	})
	if err != nil {
		return err
	}
	msg := "Successfully disabled http auth"
	if value {
		msg = "Successfully enabled http auth"
	}
	fmt.Println(msg)

	progInfo.Public = true
	if value {
		progInfo.Public = false
	}
	runtimeManager.StoreProgInfo(progInfo)
	return nil
}
