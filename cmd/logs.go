package cmd

import (
	"fmt"
	"os"
	"time"

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

	lastToken := ""

	logs := make([]api.LogType, 0)
	for {
		res, err := client.GetLogs(&api.GetLogsRequest{
			ProgramID: progInfo.ID,
			Start:     startTime,
			End:       endTime,
			LastToken: lastToken,
		})
		if err != nil {
			return err
		}
		logs = append(logs, res.Logs...)

		if len(res.LastToken) == 0 {
			break
		}

		lastToken = res.LastToken
	}

	for _, log := range logs {
		printLogs(log.Timestamp, log.Log)
	}

	return nil
}

func printLogs(timestamp int64, message string) {
	strDateTime := time.Time(time.Unix(0, timestamp*int64(time.Millisecond))).Format(time.RFC3339)
	fmt.Printf("[%s]%s", strDateTime, message)
}
