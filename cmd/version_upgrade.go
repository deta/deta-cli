package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

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
	latest, err := getLatestVersion()
	if err != nil {
		return err
	}

	upgradingTo := latest.Tag
	if versionFlag != "" {
		if !strings.HasPrefix(versionFlag, "v") {
			versionFlag = fmt.Sprintf("v%s", versionFlag)
		}

		versionExists, err := checkVersionExists(versionFlag)
		if err != nil {
			return err
		}
		if !versionExists {
			return fmt.Errorf("no such version")
		}

		upgradingTo = versionFlag
	}
	if detaVersion == upgradingTo {
		fmt.Println(fmt.Sprintf("Version already %s, no upgrade required", upgradingTo))
		return nil
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		return upgradeUnix()
	case "windows":
		return upgradeWin()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func upgradeUnix() error {
	curlCmd := exec.Command("curl", "-fsSL", "https://get.deta.dev/cli.sh")
	msg := "Upgrading deta cli"
	curlOutput, err := curlCmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(curlOutput))
		return err
	}

	co := string(curlOutput)
	shCmd := exec.Command("sh", "-c", co)
	if versionFlag != "" {
		if !strings.HasPrefix(versionFlag, "v") {
			versionFlag = fmt.Sprintf("v%s", versionFlag)
		}
		msg = fmt.Sprintf("%s to version %s", msg, versionFlag)
		shCmd = exec.Command("sh", "-c", co, "upgrade", versionFlag)
	}
	fmt.Println(fmt.Sprintf("%s...", msg))

	shOutput, err := shCmd.CombinedOutput()
	fmt.Println(string(shOutput))
	if err != nil {
		return err
	}
	return nil
}

func versionUpgradeExamples() string {
	return `
1. deta version upgrade

Upgrade cli to latest version.

2. deta version upgrade --version v1.0.0

Upgrade cli to version 'v1.0.0'.`
}
