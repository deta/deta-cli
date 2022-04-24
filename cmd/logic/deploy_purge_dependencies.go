package logic

import (
	"fmt"
	"os"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
)

func PurgeDeps(client *api.DetaClient, args []string) error {
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

	fmt.Println("Purging dependencies...")
	command, ok := runtime.DepCommands[progInfo.RuntimeName]
	if !ok {
		return fmt.Errorf("failed to determine runtime command")
	}
	purgeCmd := fmt.Sprintf("%s clean", command)
	o, err := client.UpdateProgDeps(&api.UpdateProgDepsRequest{
		ProgramID: progInfo.ID,
		Command:   purgeCmd,
	})
	if err != nil {
		return err
	}
	fmt.Println(o.Output)
	if o.HasError {
		fmt.Println()
		return fmt.Errorf("failed to purge dependencies, see output for details")
	}
	err = reloadDeps(client, runtimeManager, progInfo)
	if err != nil {
		fmt.Printf("failed to update local cached dependencies state,\nfurther calls to `deta deploy` might lead to unexpected behaviour")
	}
	return nil
}
