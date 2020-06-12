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
		Use:   "update",
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
	if newProgName == "" && envsPath == "" {
		cmd.Usage()
		return nil
	}

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
		progInfo.Name = newProgName
		runtimeManager.StoreProgInfo(progInfo)

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
		for k, v := range envChanges.Vars {
			// cant' take the address of iterated value directly
			value := v
			vars[k] = &value
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
		for k := range envChanges.Vars {
			if !inSlice(progInfo.Envs, k) {
				progInfo.Envs = append(progInfo.Envs, k)
			}
		}
		for _, d := range envChanges.Removed {
			progInfo.Envs = removeFromSlice(progInfo.Envs, d)
		}
		runtimeManager.StoreProgInfo(progInfo)

		fmt.Println("Successfully updated program environment variables.")
	}
	return nil
}
