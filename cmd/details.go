package cmd

import (
	"fmt"
	"os"

	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	detailsCmd = &cobra.Command{
		Use:   "details [flags] [path]",
		Short: "Details about a deta micro",
		RunE:  details,
		Args:  cobra.MaximumNArgs(1),
	}
)

func init() {
	rootCmd.AddCommand(detailsCmd)
}

func details(cmd *cobra.Command, args []string) error {
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
	isInitialzied, err := runtimeManager.IsInitialized()
	if err != nil {
		return err
	}
	if !isInitialzied {
		return fmt.Errorf("no deta micro initialized in '%s'", wd)
	}
	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}
	if progInfo == nil {
		return fmt.Errorf("failed to get deta micro details")
	}
	output, err := progInfoToOutput(progInfo)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
