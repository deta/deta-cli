package cmd

import (
	"fmt"
	"strings"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	forcePull bool

	pullCmd = &cobra.Command{
		Use:     "pull [flags]",
		Short:   "Pull the lastest deployed code of a deta micro",
		RunE:    pull,
		Example: pullExamples(),
		Args:    cobra.MaximumNArgs(1),
	}
)

func init() {
	pullCmd.Flags().BoolVarP(&forcePull, "force", "f", false, "force overwrite of existing files")
	rootCmd.AddCommand(pullCmd)
}

func pull(cmd *cobra.Command, args []string) error {
	runtimeManager, err := runtime.NewManager(nil, false)
	if err != nil {
		return err
	}

	isInitialized, err := runtimeManager.IsInitialized()
	if err != nil {
		return err
	}
	if !isInitialized {
		return fmt.Errorf(fmt.Sprintf("no deta micro initialized in current directory"))
	}

	if !forcePull {
		fmt.Println(fmt.Sprintf("Files already present may be overwritten. Continue? [y/n]"))
		var cont string
		fmt.Scanf("%s", &cont)
		if strings.ToLower(cont) != "y" {
			fmt.Println("Pull aborted")
			return nil
		}

	}

	progInfo, err := runtimeManager.GetProgInfo()
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

	err = runtimeManager.WriteProgramFiles(o.Files, nil, true)
	if err != nil {
		return err
	}
	err = runtimeManager.StoreProgInfo(progInfo)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fmt.Sprintf("Successfully pulled latest deployed code"))
	return nil
}

func pullExamples() string {
	return `
1. deta pull

Pull latest changes of deta micro present in the current directory. 
Asks for approval before overwriting the files in the current directory.

2. deta pull --force

Force pull latest changes of deta micro present in the current directory.
Overwrites the files present in the current directory.`
}
