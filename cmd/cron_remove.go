package cmd

import (
	"fmt"
	"os"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	cronRemoveCmd = &cobra.Command{
		Use:   "remove [path]",
		Short: "Remove a schedule from a deta micro",
		Args:  cobra.MaximumNArgs(1),
		RunE:  removeCron,
	}
)

func init() {
	cronCmd.AddCommand(cronRemoveCmd)
}

func removeCron(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("no deta micro present in '%s'", wd)
	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}

	if progInfo == nil {
		return fmt.Errorf("failed to get micro info")
	}

	err = client.DeleteSchedule(&api.DeleteScheduleRequest{
		ProgramID: progInfo.ID,
	})

	if err != nil {
		return err
	}
	fmt.Println("Successfully removed schedule from micro")

	progInfo.Cron = ""
	runtimeManager.StoreProgInfo(progInfo)
	return nil
}
