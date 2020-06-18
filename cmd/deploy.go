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
		Short: "Deploy a deta micro",
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
		return fmt.Errorf("deta micro not initilialized. see `deta new --help` to create a micro")
	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}
	c, err := runtimeManager.GetChanges()
	if err != nil {
		return err
	}

	err = reloadDeps(runtimeManager, progInfo)
	if err != nil {
		return err
	}

	dc, err := runtimeManager.GetDepChanges()
	if err != nil {
		return err
	}

	if c == nil && dc == nil {
		fmt.Println("already up to date")
		return nil
	}

	if c != nil {
		fmt.Println("Deploying...")
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

		msg := "Successfully deployed changes"
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
				return err
			}
			fmt.Println(o.Output)
			if o.HasError {
				fmt.Println()
				return fmt.Errorf("failed to update dependecies: error on one or more dependencies, no dependencies were added, see output for details")
			}
			progDetails, err := client.GetProgDetails(&api.GetProgDetailsRequest{
				ProgramID: progInfo.ID,
			})
			// if can't cet program details, set reload deps to true
			// so that it reloads the deps from the backend on next iteration
			if err != nil {
				progInfo.ReloadDeps = true
				return nil
			}

			progInfo.Deps = progDetails.Deps
			runtimeManager.StoreProgInfo(progInfo)
		}
		if len(dc.Removed) > 0 {
			err = reloadDeps(runtimeManager, progInfo)
			if err != nil {
				return err
			}
			uninstallCmd := fmt.Sprintf("%s uninstall", command)
			for _, d := range dc.Removed {
				uninstallCmd = fmt.Sprintf("%s %s", uninstallCmd, d)
			}
			o, err := client.UpdateProgDeps(&api.UpdateProgDepsRequest{
				ProgramID: progInfo.ID,
				Command:   uninstallCmd,
			})
			if err != nil {
				return err
			}
			fmt.Println(o.Output)
			if o.HasError {
				fmt.Println()
				return fmt.Errorf("failed to remove dependecies: error on one or more dependencies, no dependencies were removed, see output for details")
			}
			progDetails, err := client.GetProgDetails(&api.GetProgDetailsRequest{
				ProgramID: progInfo.ID,
			})
			// if can't get prog details set reload deps to true
			if err != nil {
				progInfo.ReloadDeps = true
				return nil
			}
			progInfo.Deps = progDetails.Deps
			runtimeManager.StoreProgInfo(progInfo)
		}
	}
	return nil
}

// reloadDeps gets program details from the server and updates the prog info deps from prog details
func reloadDeps(m *runtime.Manager, p *runtime.ProgInfo) error {
	if !p.ReloadDeps {
		return nil
	}
	progDetails, err := client.GetProgDetails(&api.GetProgDetailsRequest{
		ProgramID: p.ID,
	})
	if err != nil {
		return err
	}
	p.Deps = progDetails.Deps
	err = m.StoreProgInfo(p)
	if err != nil {
		return err
	}
	p.ReloadDeps = false
	return nil
}
