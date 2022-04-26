package cmd

import (
	"github.com/deta/deta-cli/cmd/logic"
	"github.com/spf13/cobra"
)

var (
	versionFlag string

	upgradeCmd = &cobra.Command{
		Use:     "upgrade",
		Short:   "Upgrade cli version",
		Example: versionUpgradeExamples(),
		RunE:    upgrade,
		Args:    cobra.NoArgs,
	}
)

func init() {
	upgradeCmd.Flags().StringVarP(&versionFlag, "version", "v", "", "version number")
	versionCmd.AddCommand(upgradeCmd)
}

func upgrade(cmd *cobra.Command, args []string) error {
	return logic.Upgrade(client, versionFlag, args)
}

func versionUpgradeExamples() string {
	return `
1. deta version upgrade

Upgrade cli to latest version.

2. deta version upgrade --version v1.0.0

Upgrade cli to version 'v1.0.0'.`
}
