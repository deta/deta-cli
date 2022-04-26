package logic

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/deta/deta-cli/api"
)


func Upgrade(client *api.DetaClient, versionFlag string, args []string) error {
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
		fmt.Printf("Version already %s, no upgrade required\n", upgradingTo)
		return nil
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		return upgradeUnix(versionFlag)
	case "windows":
		return upgradeWin(versionFlag)
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func upgradeUnix(versionFlag string) error {
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
	fmt.Printf("%s...\n", msg)

	shOutput, err := shCmd.CombinedOutput()
	fmt.Println(string(shOutput))
	if err != nil {
		return err
	}
	return nil
}
