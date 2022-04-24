package logic

import (
	"fmt"
	"os"

	"github.com/deta/deta-cli/api"

	"github.com/deta/deta-cli/runtime"
)

func Update(client *api.DetaClient, envsPath string, progName string, runtimeName string, args []string) error {
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