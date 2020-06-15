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
		Short: "Pull the lastest deployed code of a deta program",
		RunE:  pull,
		Args:  cobra.MaximumNArgs(1),
	}
)

func init() {
	rootCmd.AddCommand(pullCmd)
}

func pull(cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	runtimeManager, err := runtime.NewManager(&wd)
	if err != nil {
		return err
	}

	isInitialized, err := runtimeManager.IsInitialized()
	if err != nil {
		return err
	}

	if !isInitialized {
		return fmt.Errorf("no deta program initialized in '%s'", wd)
	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}

	pullPath := fmt.Sprintf("%s", progInfo.Name)
	if len(args) != 0 {
		pullPath = filepath.Join(args[0], pullPath)
	}

	_, err = os.Stat(pullPath)
	if err == nil {
		fmt.Println(fmt.Sprintf("'%s' already exists. Files already present may be overwritten. Continue? [y/n]", pullPath))
		var cont string
		fmt.Scanf("%s", &cont)
		if strings.ToLower(cont) != "y" {
			fmt.Println("Pull aborted")
			return nil
		}
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

	fmt.Println(fmt.Sprintf("Successfully pulled latest deployed code to '%s'", pullPath))
	return nil
}
