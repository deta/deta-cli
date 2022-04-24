package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"

	"github.com/spf13/cobra"
)

var (
	envsPath string

	updateCmd = &cobra.Command{
		Use:     "update",
		Short:   "Update a deta micro",
		RunE:    update,
		Example: updateExamples(),
		Args:    cobra.MaximumNArgs(1),
	}
)

func init() {
	updateCmd.Flags().StringVarP(&envsPath, "env", "e", "", "path to env file")
	updateCmd.Flags().StringVarP(&progName, "name", "n", "", "new name of the micro")
	updateCmd.Flags().StringVarP(&runtimeName, "runtime", "r", "", "runtime version\n\tPython: python3.7, python3.8, python3.9\n\tNode: nodejs12, nodejs14")
	rootCmd.AddCommand(updateCmd)
}

func update(cmd *cobra.Command, args []string) error {
	if len(progName) == 0 && len(envsPath) == 0 && len(runtimeName) == 0 {
		cmd.Usage()
		return nil
	}

	return logic.Update(client, envsPath, progName, runtimeName, args)
}

func updateExamples() string {
	return `
1. deta update --name a-new-name

Update the name of a deta micro with a new name "a-new-name".

2. deta update --env env-file

Update the enviroment variables of a deta micro from the file 'env-file'. 
File 'env-file' must have env vars of format 'key=value'.

3. deta update --runtime nodejs12

Update the runtime of a deta micro.
Available runtimes:
	Python: python3.7, python3.8, python3.9
	Node: nodejs12, nodejs14`
}
