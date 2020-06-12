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
		Short: "Deploy a program",
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

	dc, err := runtimeManager.GetDepChanges()
	if err != nil {
		return err
	}

	if c == nil && dc == nil {
		fmt.Println("Program already up to date.")
		return nil
	}

	if c != nil {
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

		msg := "Successfully deployed code changes."
		fmt.Println(msg)
		runtimeManager.StoreState()
	}

	if dc != nil {
		fmt.Println("Updating dependencies...")
		command := runtime.DepCommands[progInfo.Runtime]
		if len(dc.Added) > 0 {
			installCmd := fmt.Sprintf("%s install", command)
			for _, a := range dc.Added {
				installCmd = fmt.Sprintf("%s %s", installCmd, a)
			}
			o, err := client.UpdateProgDeps(&api.UpdateProgDepsRequest{
				ProgramID: progInfo.ID,
				Command:   installCmd,
			})
			if err != nil {
				return fmt.Errorf("failed to add dependencies: %v", err)
			}
			fmt.Println(o.Output)

			for _, a := range dc.Added {
				progInfo.Deps = append(progInfo.Deps, a)
			}
			runtimeManager.StoreProgInfo(progInfo)
		}
		if len(dc.Removed) > 0 {
			uninstallCmd := fmt.Sprintf("%s uninstall", command)
			for _, d := range dc.Removed {
				uninstallCmd = fmt.Sprintf("%s %s", uninstallCmd, d)
			}
			o, err := client.UpdateProgDeps(&api.UpdateProgDepsRequest{
				ProgramID: progInfo.ID,
				Command:   uninstallCmd,
			})
			if err != nil {
				return fmt.Errorf("failed to remove dependencies: %v", err)
			}
			fmt.Println(o.Output)
			for _, d := range dc.Removed {
				progInfo.Deps = removeFromSlice(progInfo.Deps, d)
			}
			runtimeManager.StoreProgInfo(progInfo)
		}
	}
	return nil
}
