package cmd

import (
	"fmt"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "deploy a program",
		Args:  cobra.MaximumNArgs(1),
		RunE:  deploy,
	}
)

func init() {
	rootCmd.AddCommand(deployCmd)
}

func deploy(cmd *cobra.Command, args []string) error {
	var err error
	var runtimeManager *runtime.Manager

	if len(args) == 0 {
		// new manager sets curent working directory as root directory
		// if root is not provided
		runtimeManager, err = runtime.NewManager(nil)
	} else {
		runtimeManager, err = runtime.NewManager(&args[0])
	}
	if err != nil {
		return err
	}

	isInitialized, err := runtimeManager.IsInitialized()
	if err != nil {
		return err
	}

	if !isInitialized {
		return fmt.Errorf("deta program not initilialized. see `deta new --help` to create a program")
	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}
	c, err := runtimeManager.GetChanges()
	if err != nil {
		return err
	}
	if c == nil {
		fmt.Println("No changes to be deployed")
		return nil
	}

	_, err = client.Deploy(&api.DeployRequest{
		ProgramID: progInfo.ID,
		Changes:   c.Changes,
		Deletions: c.Deletions,
		Account:   progInfo.Account,
		Region:    progInfo.Region,
	})
	if err != nil {
		return err
	}

	runtimeManager.StoreState()
	return nil
}
