package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
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
	return logic.SetCron(client, args)
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
