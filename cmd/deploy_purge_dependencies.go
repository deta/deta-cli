package cmd

import (
	"fmt"
	"os"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	purgeDepsCmd = &cobra.Command{
		Use:     "purge-dependencies [path]",
		Short:   "Remove all installed dependencies for a deta micro",
		Args:    cobra.MaximumNArgs(1),
		Example: purgeDepsExamples(),
		RunE:    purgeDeps,
	}
)

func init() {
	deployCmd.AddCommand(purgeDepsCmd)
}

func purgeDeps(cmd *cobra.Command, args []string) error {
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
	err = reloadDeps(runtimeManager, progInfo)
	if err != nil {
		fmt.Printf("failed to update local cached dependencies state,\nfurther calls to `deta deploy` might lead to unexpected behaviour")
	}
	return nil
}

func purgeDepsExamples() string {
	return `
1. deta deploy purge-dependencies	

Remove all dependencies installed for the micro in the current directory

2. deta deploy purge-dependencies ./my-micro

Remove all dependencies installed for the micro in './my-micro' directory`
}
