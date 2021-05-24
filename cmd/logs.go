package cmd

import (
	"fmt"
	"os"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

const (
	SPACE = " "
)

var (
	start   string
	end     string
	logsCmd = &cobra.Command{
		Use:   "logs [flags]",
		Short: "Get logs from micro",
		Args:  cobra.NoArgs,
		RunE:  logs,
	}
)

func init() {
	logsCmd.Flags().StringVar(&start, "start", "", "logs start time 'format: RFC3339'")
	logsCmd.Flags().StringVar(&end, "end", "", "logs end time 'format: RFC3339'")

	rootCmd.AddCommand(logsCmd)
}

func logs(cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	runtimeManager, err := runtime.NewManager(&wd, false)
	if err != nil {
		return nil
	}

	isInitialized, err := runtimeManager.IsInitialized()
	if err != nil {
		return err
	}

	if !isInitialized {
		return fmt.Errorf("no deta micro initialized in '%s'", wd)
	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}

	if progInfo == nil {
		return fmt.Errorf("failed to get micro information")
	}

	startTime, err := parseDateTime(start)
	if err != nil {
		return err
	}

	endTime, err := parseDateTime(end)
	if err != nil {
		return err
	}

	res, err := client.GetLogs(&api.GetLogsRequest{
		ProgramID: progInfo.ID,
		Start:     startTime,
		End:       endTime,
	})
	if err != nil {
		return err
	}

	fmt.Println(res.Logs)

	return nil
}
