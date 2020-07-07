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
	inputFile string
	runCmd    = &cobra.Command{
		Use:   "run [flags] [-- [action] <input>]",
		Short: "Run a deta micro",
		RunE:  run,
	}
)

func init() {
	rootCmd.AddCommand(runCmd)
}

func run(cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()

	runtimeManager, err := runtime.NewManager(&wd, false)
	if err != nil {
		return nil
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
		return fmt.Errorf(fmt.Sprintf("failed to get micro information"))
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
	res, err := client.InvokeProgram(req)
	if err != nil {
		return err
	}

	out, err := prettyPrint(res)
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
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
					switch v.(type) {
					case string:
						progInput[key] = []string{v.(string), value}
					case []string:
						progInput[key] = append(v.([]string), value)
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
