package logic

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
)

var (
	errInvalidExp = errors.New("invalid expression")
)

func SetCron(client *api.DetaClient, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	var expr string
	if len(args) == 2 {
		wd = args[0]
		expr = args[1]
	} else if len(args) == 1 {
		expr = args[0]
	} else {
		return fmt.Errorf("no expression provided")
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

	cronType, err := getCronTypeFromExpr(expr)
	if err != nil {
		return fmt.Errorf("invalid cron expression, see `deta cron set --help`")
	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}

	if progInfo == nil {
		return fmt.Errorf("failed to get micro info")
	}

	fmt.Println("Scheduling micro...")
	err = client.AddSchedule(&api.AddScheduleRequest{
		ProgramID:  progInfo.ID,
		Type:       cronType,
		Expression: expr,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Successfully set micro to schedule for '%s'\n", expr)

	progInfo.Cron = expr
	runtimeManager.StoreProgInfo(progInfo)
	return nil
}

func getCronTypeFromExpr(expr string) (string, error) {
	parts := strings.Split(expr, " ")
	if len(parts) == 2 {
		return "rate", nil
	}
	if len(parts) == 6 {
		return "cron", nil
	}
	return "", errInvalidExp
}