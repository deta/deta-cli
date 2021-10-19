package cmd

import (
	"fmt"
	"os"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	deployCmd = &cobra.Command{
		Use:     "deploy [path]",
		Short:   "Deploy a deta micro",
		Args:    cobra.MaximumNArgs(1),
		Example: deployExamples(),
		RunE:    deploy,
	}
)

func init() {
	rootCmd.AddCommand(deployCmd)
}

func deploy(cmd *cobra.Command, args []string) error {
	// check version
	c := make(chan *checkVersionMsg, 1)
	defer close(c)
	go checkVersion(c)

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
		return fmt.Errorf("no deta micro initilialized, see `deta new --help` to create a micro")
	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}
	err = deployChanges(runtimeManager, progInfo, false)
	if err != nil {
		return err
	}
	cm := <-c
	if cm.err == nil && cm.isLower {
		fmt.Println("New Deta CLI version available, upgrade with `deta version upgrade`")
	}
	return nil
}

// reloadDeps gets program details from the server and updates the prog info deps from prog details
func reloadDeps(m *runtime.Manager, p *runtime.ProgInfo) error {
	progDetails, err := client.GetProgDetails(&api.GetProgDetailsRequest{
		Program: p.ID,
		Project: p.Project,
		Space:   p.Space,
	})
	if err != nil {
		return err
	}
	p.Deps = progDetails.Deps
	err = m.StoreProgInfo(p)
	if err != nil {
		return err
	}
	return nil
}

func deployChanges(m *runtime.Manager, p *runtime.ProgInfo, isWatcher bool) error {
	c, err := m.GetChanges()
	if err != nil {
		return err
	}

	err = reloadDeps(m, p)
	if err != nil {
		return err
	}

	dc, err := m.GetDepChanges()
	if err != nil {
		return err
	}

	if c == nil && dc == nil {
		// workaround for multiple write events fired
		// with file watcher
		if !isWatcher {
			fmt.Println("Everything up to date")
		}
		return nil
	}

	if c != nil {
		fmt.Println("Deploying...")
		_, err = client.Deploy(&api.DeployRequest{
			ProgramID:   p.ID,
			Changes:     c.Changes,
			Deletions:   c.Deletions,
			BinaryFiles: c.BinaryFiles,
			Account:     p.Account,
			Region:      p.Region,
		})
		if err != nil {
			return err
		}

		msg := "Successfully deployed changes"
		fmt.Println(msg)
		m.StoreState()
	}

	if dc != nil {
		fmt.Println("Updating dependencies...")
		command := runtime.DepCommands[p.RuntimeName]
		if len(dc.Removed) > 0 {
			uninstallCmd := ""
			// clean all deps if everything is removed
			if areSlicesEqualNoOrder(dc.Removed, p.Deps) {
				uninstallCmd = fmt.Sprintf("%s clean", command)
			} else {
				uninstallCmd = fmt.Sprintf("%s uninstall", command)
				for _, d := range dc.Removed {
					uninstallCmd = fmt.Sprintf("%s %s", uninstallCmd, d)
				}
			}
			o, err := client.UpdateProgDeps(&api.UpdateProgDepsRequest{
				ProgramID: p.ID,
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
		}
		if len(dc.Added) > 0 {
			installCmd := fmt.Sprintf("%s install", command)
			for _, a := range dc.Added {
				installCmd = fmt.Sprintf("%s %s", installCmd, a)
			}
			o, err := client.UpdateProgDeps(&api.UpdateProgDepsRequest{
				ProgramID: p.ID,
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
		}
		err = reloadDeps(m, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func deployExamples() string {
	return `
1. deta deploy

Deploy a deta micro rooted in the current directory.

2. deta deploy micros/my-micro-1

Deploy a deta micro rooted in 'micros/my-micro-1' directory.`
}
