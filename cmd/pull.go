package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	pullCmd = &cobra.Command{
		Use:   "pull [path_to_pull_to]",
		Short: "Pull the lastest deployed code of a deta micro",
		RunE:  pull,
		Args:  cobra.MaximumNArgs(1),
	}
)

func init() {
	newCmd.Flags().StringVarP(&progName, "n", "name", "", "name of the micro")
	newCmd.Flags().StringVarP(&projectName, "p", "project", "default", "project of the micro")
	newCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(pullCmd)
}

func pull(cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if len(args) > 0 {
		wd = args[0]
	}

	pullPath := filepath.Join(wd, progName)
	i, err := os.Stat(wd)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(pullPath, 0760)
			if err != nil {
				return err
			}
		}
		return err
	}
	if !i.IsDir() {
		return fmt.Errorf("'%s' is not a directory", wd)
	}
	if _, err := os.Stat(pullPath); err == nil {
		fmt.Println(fmt.Sprintf("'%s' already exists. Files already present may be overwritten. Continue? [y/n]", pullPath))
		var cont string
		fmt.Scanf("%s", &cont)
		if strings.ToLower(cont) != "y" {
			fmt.Println("Pull aborted")
			return nil
		}
	}

	runtimeManager, err := runtime.NewManager(&wd)
	if err != nil {
		return err
	}

	isInitialized, err := runtimeManager.IsInitialized()
	if err != nil {
		return err
	}

	if isInitialized {
		return fmt.Errorf(fmt.Sprintf("another deta micro already present in '%s'", wd))
	}

	u, err := runtimeManager.GetUserInfo()
	if err != nil {
		return err
	}

	if u == nil {
		fmt.Println("Login required. Please, log in with `deta login`")
		return nil
	}

	progDetails, err := client.GetProgDetails(&api.GetProgDetailsRequest{
		Program: progName,
		Project: projectName,
		Space:   u.DefaultSpace,
	})
	if err != nil {
		return err
	}

	progInfo := &runtime.ProgInfo{
		ID:      progDetails.ID,
		Space:   progDetails.Space,
		Runtime: progDetails.Runtime,
		Name:    progDetails.Name,
		Path:    progDetails.Path,
		Project: progDetails.Project,
		Account: progDetails.Account,
		Region:  progDetails.Region,
		Deps:    progDetails.Deps,
		Envs:    progDetails.Envs,
		Public:  progDetails.Public,
	}

	runtimeManager, err = runtime.NewManager(&wd)
	if err != nil {
		return err
	}

	o, err := client.DownloadProgram(&api.DownloadProgramRequest{
		ProgramID: progInfo.ID,
		Runtime:   progInfo.Runtime,
		Account:   progInfo.Account,
		Region:    progInfo.Region,
	})

	if err != nil {
		return err
	}

	err = runtimeManager.WriteProgramFiles(o.Files, &pullPath)
	if err != nil {
		return err
	}
	runtimeManager.StoreProgInfo(progInfo)
	fmt.Println(fmt.Sprintf("Successfully pulled latest deployed code to '%s'", pullPath))
	return nil
}
