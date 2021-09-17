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
		Use:     "update",
		Short:   "Update a deta micro",
		RunE:    update,
		Example: updateExamples(),
		Args:    cobra.MaximumNArgs(1),
	}
)

func init() {
	updateCmd.Flags().StringVarP(&envsPath, "env", "e", "", "path to env file")
	updateCmd.Flags().StringVarP(&progName, "name", "n", "", "new name of the micro")
	updateCmd.Flags().StringVarP(&runtimeName, "runtime", "r", "", "runtime version\n\tPython: python3.7, python3.9\n\tNode: nodejs12, nodejs14")
	rootCmd.AddCommand(updateCmd)
}

func update(cmd *cobra.Command, args []string) error {
	if len(progName) == 0 && len(envsPath) == 0 && len(runtimeName) == 0 {
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

	runtimeManager, err := runtime.NewManager(&wd, false)
	if err != nil {
		return err
	}

	isInitialized, err := runtimeManager.IsInitialized()
	if err != nil {
		return err
	}
	if !isInitialized {
		return fmt.Errorf("no deta micro initialized in `%s'", wd)
	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}

	if len(progName) != 0 {
		fmt.Println("Updating the name...")
		err := client.UpdateProgName(&api.UpdateProgNameRequest{
			ProgramID: progInfo.ID,
			Name:      progName,
		})
		if err != nil {
			return err
		}
		progInfo.Name = progName
		runtimeManager.StoreProgInfo(progInfo)

		fmt.Println("Successfully update micro's name")
	}

	if len(envsPath) != 0 {
		fmt.Println("Updating environment variables...")
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
			return err
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

		fmt.Println("Successfully updated micro's environment variables")
	}

	if len(runtimeName) != 0 {
		fmt.Println("Updating runtime...")
		progRuntime, err := parseRuntime(runtimeName)
		if err != nil {
			return err
		}

		err = client.UpdateProgRuntime(&api.UpdateProgRuntimeRequest{
			ProgramID: progInfo.ID,
			Runtime:   progRuntime.Version,
		})
		if err != nil {
			return err
		}
		progInfo.Runtime = progRuntime.Version
		runtimeManager.StoreProgInfo(progInfo)

		fmt.Println("Successfully update micro's runtime")
	}

	return nil
}

func updateExamples() string {
	return `
1. deta update --name a-new-name

Update the name of a deta micro with a new name "a-new-name".

2. deta update --env env-file

Update the enviroment variables of a deta micro from the file 'env-file'. 
File 'env-file' must have env vars of format 'key=value'.

3. deta update --runtime nodejs12

Update the runtime of a deta micro.
Available runtimes:
	Python: python3.7, python3.9
	Node: nodejs12, nodejs14`
}
