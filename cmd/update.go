package cmd

import (
	"fmt"
	"os"

	"github.com/deta/deta-cli/api"

	"github.com/deta/deta-cli/runtime"

	"github.com/spf13/cobra"
)

var (
	envsPath string

	updateCmd = &cobra.Command{
		Use:   "update [--name name] [--envs env_file_path]",
		Short: "Update a program",
		RunE:  update,
		Args:  cobra.MaximumNArgs(1),
	}
)

func init() {
	updateCmd.Flags().StringVarP(&envsPath, "env", "e", "", "path to env file")
	updateCmd.Flags().StringVarP(&newProgName, "name", "n", "", "new name of the program")
	rootCmd.AddCommand(updateCmd)
}

func update(cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if len(args) != 0 {
		wd = args[0]
	}

	runtimeManager, err := runtime.NewManager(&wd)
	if err != nil {
		return err
	}

	isInitialized, err := runtimeManager.IsInitialized()
	if !isInitialized {
		return fmt.Errorf("No deta program initialized in `%s'", wd)
	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}

	if newProgName != "" {
		err := client.UpdateProgName(&api.UpdateProgNameRequest{
			ProgramID: progInfo.ID,
			Name:      newProgName,
		})
		if err != nil {
			return fmt.Errorf("failed to update program: %v", err)
		}
		msg := "Successfully updated program name."
		if envsPath != "" {
			msg = fmt.Sprintf("%s%s", msg, "Updating environment variables...")
		}
		fmt.Println(msg)
	}

	if envsPath != "" {
		envChanges, err := runtimeManager.GetEnvChanges(envsPath)
		if err != nil {
			return fmt.Errorf("failed to update env vars: %v", err)
		}
		vars := make(map[string]*string)
		for a, v := range envChanges.Added {
			vars[a] = &v
		}
		for _, d := range envChanges.Removed {
			vars[d] = nil
		}

		err = client.UpdateProgEnvs(&api.UpdateProgEnvsRequest{
			ProgramID: progInfo.ID,
			Account:   progInfo.Account,
			Region:    progInfo.Region,
			Vars:      vars,
		})
		if err != nil {
			return fmt.Errorf("Failed to update program environment variables: %v", err)
		}
		fmt.Println("Successfully update program environment variables.")
	}
	return nil
}
