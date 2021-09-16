// +build windows

package cmd

import (
	"fmt"
	"os/exec"
)

func upgradeWin() error {
	msg := "Upgrading deta cli"
	cmd := "iwr https://get.deta.dev/cli.ps1 -useb | iex"

	if versionFlag != "" {
		msg = fmt.Sprintf("%s to version %s", msg, versionFlag)
		cmd = fmt.Sprintf(`$v="%s"; %s`, versionFlag, cmd)
	}
	fmt.Println(fmt.Sprintf("%s...", msg))

	pshellCmd := exec.Command("powershell", cmd)

	stdout, err := pshellCmd.CombinedOutput()
	fmt.Println(string(stdout))
	if err != nil {
		return err
	}

	return nil
}
