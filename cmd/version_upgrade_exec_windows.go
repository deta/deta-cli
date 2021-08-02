// +build windows

package cmd

import (
	"fmt"
	"strings"

	ps "github.com/bhendo/go-powershell"
	"github.com/bhendo/go-powershell/backend"
)

func upgradeWin() error {
	back := &backend.Local{}

	shell, err := ps.New(back)
	if err != nil {
		return err
	}

	defer shell.Exit()

	msg := "Upgrading deta cli"
	cmd := "iwr https://get.deta.dev/cli.ps1 -useb | iex"
	if versionFlag != "" {
		msg = fmt.Sprintf("%s to version %s", msg, versionFlag)
		if strings.HasPrefix(versionFlag, "v") {
			versionFlag = versionFlag[1:]
		}
		cmd = fmt.Sprintf(`$v="%s"; %s`, versionFlag, cmd)
	}
	fmt.Println(fmt.Sprintf("%s...", msg))

	stdout, stderr, err := shell.Execute(cmd)
	if err != nil {
		fmt.Println(stderr)
		return err
	}
	fmt.Println(stdout)
	return nil
}
