package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/deta/deta-cli/api"
	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	errInvalidExp = errors.New("invalid expression")

	cronSetCmd = &cobra.Command{
		Use:     "set [path] <expression>",
		Short:   "Set deta micro to run on a schedule",
		Args:    cobra.MaximumNArgs(2),
		Example: setExamples(),
		RunE:    setCron,
	}
)

func init() {
	cronCmd.AddCommand(cronSetCmd)
}

func setCron(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("no deta micro presentin '%s'", wd)
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
	fmt.Println(fmt.Sprintf(`Successfully set micro to schedule for "%s"`, expr))

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

func setExamples() string {
	return `
Rate:

1. deta cron set "1 minute" : run every minute
2. deta cron set "5 hours" : run every five hours

Cron expressions:

1. deta cron set "0 10 * * ? *" : run at 10:00 am(UTC) every day
2. deta cron set "30 18 ? * MON-FRI *" : run at 6:00 pm(UTC) Monday through Friday
3. deta cron set "0/5 8-17 ? * MON-FRI *" : run every 5 minutes Monday through Friday between 8:00 am and 5:55 pm(UTC)

See more examples at https://docs.deta.sh`
}
