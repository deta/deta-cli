package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

const (
	TickerDuration = 250
)

var (
	debugCmd = &cobra.Command{
		Use:   "debug [path]",
		Short: "debug logs in real time",
		Args:  cobra.MaximumNArgs(1),
		RunE:  debug,
	}
)

func init() {
	rootCmd.AddCommand(debugCmd)
}

func debug(cmd *cobra.Command, args []string) error {

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	start := time.Now().UnixNano() / int64(time.Millisecond)

	ticker := time.NewTicker(TickerDuration * time.Millisecond)
	defer ticker.Stop()

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
		return fmt.Errorf("no deta micro initialized in '%s'", wd)
	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}

	if progInfo == nil {
		return fmt.Errorf("failed to get micro information")
	}

	enableVisor := false
	if progInfo.Visor == "debug" {
		err = updateVisor("off", args)
		if err != nil {
			return err
		}
		enableVisor = true
	}
	fmt.Println("Listening for logs")

	go func() {
		for {
			time.Sleep(TickerDuration * time.Millisecond)
		}
	}()

	for {
		select {
		case <-sigs:
			if enableVisor {
				_ = updateVisor("debug", args)
			}
			return nil
		case <-ticker.C:
			_ = pollLogs(progInfo, start)
		}
	}
}

func pollLogs(progInfo *runtime.ProgInfo, start int64) error {
	lk := make(map[int64]struct{})
	logs := make([]api.LogType, 0)
	lastToken := ""
	for {
		res, err := client.GetLogs(&api.GetLogsRequest{
			ProgramID: progInfo.ID,
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
		_, ok := lk[log.Timestamp]
		if !ok && log.Timestamp > start {
			printLogs(log.Timestamp, log.Log)
			lk[log.Timestamp] = struct{}{}
		}
	}
	return nil
}
