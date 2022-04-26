package logic

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
)

const (
	SPACE                    = " "
	LogPollDurationInSeconds = 1
)


func Logs(client *api.DetaClient, followFlag bool, args []string) error {
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

	// no follow flag simply print the logs
	if !followFlag {
		logs, err := getLogs(client, progInfo.ID)
		if err != nil {
			return err
		}

		for _, log := range logs {
			printLogs(log.Timestamp, log.Log)
		}

		return nil
	}

	// follow flag specified
	return followLogs(client, progInfo)
}

func getLogs(client *api.DetaClient, progID string) ([]api.LogType, error) {
	lastToken := ""
	logs := make([]api.LogType, 0)
	current := time.Now().UTC()
	for {
		res, err := client.GetLogs(&api.GetLogsRequest{
			ProgramID: progID,
			Start:     current.Add(-30*time.Minute).UnixNano() / int64(time.Millisecond),
			End:       current.UnixNano() / int64(time.Millisecond),
			LastToken: lastToken,
		})
		if err != nil {
			return nil, err
		}
		logs = append(logs, res.Logs...)

		if len(res.LastToken) == 0 {
			break
		}
		lastToken = res.LastToken
	}
	return logs, nil
}

// show new logs only shows logs after start and not in seenLogs
func showNewLogs(client *api.DetaClient, progID string, start int64, seenLogs map[int64]struct{}) error {
	logs, err := getLogs(client, progID)
	if err != nil {
		return err
	}

	for _, log := range logs {
		_, ok := seenLogs[log.Timestamp]
		if !ok && log.Timestamp > start {
			printLogs(log.Timestamp, log.Log)
			seenLogs[log.Timestamp] = struct{}{}
		}
	}
	return nil
}

// follow logs polls for new logs
// waits on a poll ticker or a signal
func followLogs(client *api.DetaClient, progInfo *runtime.ProgInfo) error {
	start := time.Now().UTC().UnixNano() / int64(time.Millisecond)

	// signals channel
	sigs := make(chan os.Signal, 1)

	// notify on Ctrl+C, Ctrl+D, Terminate, Quit
	signal.Notify(sigs,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	fmt.Println("Listening for logs...")
	// disable visor mode temporarily if it's on
	enableVisor := false
	if progInfo.Visor == "debug" {
		err := client.UpdateVisorMode(&api.UpdateVisorModeRequest{
			ProgramID: progInfo.ID,
			Mode:      "off",
		})
		if err != nil {
			return err
		}
		enableVisor = true
	}

	// log poll ticker
	ticker := time.NewTicker(LogPollDurationInSeconds * time.Second)
	defer ticker.Stop()

	// track seen logs
	seenLogs := make(map[int64]struct{})
	for {
		select {
		case <-sigs:
			ticker.Stop()
			if !enableVisor {
				return nil
			}
			// renable visor if visor was on
			err := client.UpdateVisorMode(&api.UpdateVisorModeRequest{
				ProgramID: progInfo.ID,
				Mode:      "debug",
			})
			if err != nil {
				return fmt.Errorf("failed to renable visor, please renable with `deta visor enable`")
			}
			return nil
		case <-ticker.C:
			if err := showNewLogs(client, progInfo.ID, start, seenLogs); err != nil {
				return err
			}
		}
	}
}

func printLogs(timestamp int64, message string) {
	strDateTime := time.Time(time.Unix(0, timestamp*int64(time.Millisecond))).Format(time.RFC3339)
	fmt.Printf("[%s] %s\n", strDateTime, message)
}
