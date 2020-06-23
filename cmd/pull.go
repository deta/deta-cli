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
		Use:   "pull [flags] [path]",
		Short: "Pull the lastest deployed code of a deta micro",
		RunE:  pull,
		Args:  cobra.MaximumNArgs(1),
	}
)

func init() {
	pullCmd.Flags().StringVar(&progName, "name", "", "deta micro name")
	pullCmd.Flags().StringVar(&projectName, "project", "", "deta project")
	pullCmd.MarkFlagRequired("name")

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

	runtimeManager, err := runtime.NewManager(&pullPath, true)
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
		fmt.Println("login required, log in with `deta login`")
		return nil
	}

	if projectName == "" {
		projectName = u.DefaultProject
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

	o, err := client.DownloadProgram(&api.DownloadProgramRequest{
		ProgramID: progInfo.ID,
		Runtime:   progInfo.Runtime,
		Account:   progInfo.Account,
		Region:    progInfo.Region,
	})

	if err != nil {
		return err
	}

	err = runtimeManager.WriteProgramFiles(o.Files, &pullPath, false)
	if err != nil {
		return err
	}
	err = runtimeManager.StoreProgInfo(progInfo)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fmt.Sprintf("Successfully pulled latest deployed code to '%s'", pullPath))
	return nil
}
