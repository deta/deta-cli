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
	N = 1
)

var (
	debugCmd = &cobra.Command{
		Use:   "debug [flags]",
		Short: "debug logs in real time",
		Args:  cobra.NoArgs,
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

	enableVisor := false
	if progInfo.Visor == "debug" {
		err = updateVisor("off", args)
		if err != nil {
			return err
		}
		enableVisor = true
	}

	go func() {
		<-sigs
		if enableVisor {
			_ = updateVisor("debug", args)
		}
		os.Exit(0)
	}()

	logs := make(map[int64]struct{})
	lastToken := ""

	fmt.Println("Listening for logs")

	for {
		time.Sleep(N * time.Second)
		res, err := client.GetLogs(&api.GetLogsRequest{
			ProgramID: progInfo.ID,
			LastToken: lastToken,
		})
		if err != nil {
			return err
		}

		for _, log := range res.Logs {
			_, ok := logs[log.Timestamp]
			if !ok && log.Timestamp > start {
				printLogs(log.Timestamp, log.Log)
				logs[log.Timestamp] = struct{}{}
			}
		}
	}
}
