package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	nodeFlag    bool
	pythonFlag  bool
	progName    string
	projectName string
	runtimeName string

	newCmd = &cobra.Command{
		Use:     "new [flags] [path]",
		Short:   "Create a new deta micro",
		RunE:    new,
		Example: newExamples(),
		Args:    cobra.MaximumNArgs(1),
	}
)

func init() {
	// flags
	newCmd.Flags().BoolVarP(&nodeFlag, "node", "n", false, "create a micro with node runtime")
	newCmd.Flags().BoolVarP(&pythonFlag, "python", "p", false, "create a micro with python runtime")
	newCmd.Flags().StringVar(&progName, "name", "", "deta micro name")
	newCmd.Flags().StringVar(&projectName, "project", "", "project to create the micro under")
	newCmd.Flags().StringVar(&runtimeName, "runtime", "", "runtime version\n\tPython: python3.7, python3.8, python3.9\n\tNode: nodejs12, nodejs14")

	rootCmd.AddCommand(newCmd)
}

func new(cmd *cobra.Command, args []string) error {
	return logic.NewProgram(client, progName, projectName, runtimeName, pythonFlag, nodeFlag, args)
}

func newExamples() string {
	return `
1. deta new

Create a new deta micro from the current directory with an entrypoint file (either 'main.py' or 'index.js') already present in the directory.

2. deta new my-micro

Create a new deta micro from './my-micro' directory with an entrypoint file (either 'main.py' or 'index.js') already present in the directory.

2. deta new --node my-node-micro

Create a new deta micro with the node runtime in the directory './my-node-micro'.
'./my-node-micro' must not contain a python entrypoint file ('main.py') if directory is already present. 

3. deta new --python --name my-github-webhook webhooks/github-deta

Create a new deta micro with the python runtime, name 'my-github-webhook' and in directory 'webhooks/github-deta'. 
'webhooks/github-deta' must not contain a node entrypoint file ('index.js') if directory is already present.

4. deta new --runtime nodejs12 --name my-node-micro

Create a new deta micro with the node (nodejs12.x) runtime in the directory './my-node-micro'.
'./my-node-micro' must not contain a python entrypoint file ('main.py') if directory is already present. `
}
