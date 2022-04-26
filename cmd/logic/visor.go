package logic

import (
	"fmt"
	"os"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
)

func UpdateVisor(client *api.DetaClient, mode string, args []string) error {
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

	err = client.UpdateVisorMode(&api.UpdateVisorModeRequest{
		ProgramID: progInfo.ID,
		Mode:      mode,
	})
	if err != nil {
		return err
	}
	msg := "Successfully disabled visor mode"
	if mode == "debug" {
		msg = "Successfully enabled visor mode"
	}
	fmt.Println(msg)

	progInfo.Visor = "off"
	if mode == "debug" {
		progInfo.Visor = "debug"
	}
	runtimeManager.StoreProgInfo(progInfo)
	return nil
}
