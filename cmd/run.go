package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/deta/deta-cli/api"

	"github.com/deta/deta-cli/runtime"
	"github.com/spf13/cobra"
)

var (
	showLogs bool
	runCmd   = &cobra.Command{
		Use:     "run [flags] [action] [-- <input args>]",
		Short:   "Run a deta micro",
		Example: runExamples(),
		RunE:    run,
	}
)

func init() {
	runCmd.Flags().BoolVarP(&showLogs, "logs", "l", false, "show micro logs")
	rootCmd.AddCommand(runCmd)
}

func run(cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
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
		return fmt.Errorf(fmt.Sprintf("No deta micro present in '%s'", wd))
	}

	progInfo, err := runtimeManager.GetProgInfo()
	if err != nil {
		return err
	}

	if progInfo == nil {
		return fmt.Errorf("failed to get micro information")
	}

	action, progArgs := parseArgs(args)

	body, err := json.Marshal(progArgs)
	if err != nil {
		return err
	}

	req := &api.InvokeProgRequest{
		ProgramID: progInfo.ID,
		Action:    action,
		Body:      string(body),
	}

	fmt.Println("Running micro...")
	fmt.Println()
	res, err := client.InvokeProgram(req)
	if err != nil {
		return err
	}
	return printResponse(res.Payload, res.Logs)
}

func parseArgs(args []string) (string, map[string]interface{}) {
	var action string
	progInput := make(map[string]interface{})
	for i := 0; i < len(args); i++ {
		if i == 0 && !strings.HasPrefix(args[i], "-") {
			action = args[i]
		}
		if strings.HasPrefix(args[i], "--") {
			j := i + 1
			if key := cleanFlag(args[i]); key != "" {
				var value string
				if j < len(args) && !strings.HasPrefix(args[j], "-") {
					value = args[j]
					i = j
				}
				if v, ok := progInput[key]; ok {
					switch vt := v.(type) {
					case string:
						progInput[key] = []string{vt, value}
					case []string:
						progInput[key] = append(vt, value)
					}
				} else {
					progInput[key] = value
				}
			}
		}
		if strings.HasPrefix(args[i], "-") && !strings.HasPrefix(args[i], "--") {
			progInput[strings.TrimPrefix(args[i], "-")] = true
		}
	}
	return action, progInput
}

func cleanFlag(flag string) string {
	for i, c := range flag {
		if string(c) != "-" {
			return flag[i:]
		}
	}
	return ""
}

func cleanLogs(logs string) string {
	logsParts := strings.Split(logs, "\n")
	logsParts = logsParts[1 : len(logsParts)-3]
	return strings.Join(logsParts, "\n")
}

func printResponse(payload, logs string) error {
	var p map[string]interface{}
	err := json.Unmarshal([]byte(payload), &p)
	if err != nil {
		return err
	}

	fmt.Println("Response:")
	if b, ok := p["body"]; ok {
		var body interface{}
		err := json.Unmarshal([]byte(b.(string)), &body)
		if err != nil {
			return err
		}
		o, err := prettyPrint(body)
		if err != nil {
			return err
		}
		fmt.Println(o)
	} else {
		o, err := prettyPrint(p)
		if err != nil {
			return err
		}
		fmt.Println(o)
	}

	if showLogs {
		fmt.Println()
		fmt.Println("Logs:")
		fmt.Println(cleanLogs(logs))
	}
	return nil
}

func runExamples() string {
	return `
1. deta run -- --name Jimmy --age 33 -active

Run deta micro with the following input:
{
	"name": "Jimmy",
	"age": "33",
	"active": true
}

2. deta run --logs test -- --username admin

Run deta micro and show micro logs with action 'test' and the following input:
{
	"username": "admin"
}

3. deta run delete -- --emails jimmy@deta.sh --emails joe@deta.sh

Run deta micro with action 'delete' and the following input:
{
	"emails": ["jimmy@deta.sh", "joe@deta.sh"]
}  

See https://docs.deta.sh for more examples and details. 
`
}
