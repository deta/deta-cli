package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/spf13/cobra"
)

var (
	loginCmd = &cobra.Command{
		Use:   "login",
		Short: "login to deta",
		RunE:  login,
	}
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

func login(cmd *cobra.Command, args []string) error {
	u, p := promptCreds()
	fmt.Println(u, p)
	return nil
}

func promptCreds() (string, string) {
	fmt.Print("Username: ")
	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	if err != nil {
		os.Stderr.WriteString("Failed to read username")
		os.Exit(1)
	}
	username = strings.TrimSuffix(username, "\n")

	fmt.Print("Password: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		os.Stderr.WriteString("Failed to read password")
		os.Exit(1)
	}
	return username, string(password)
}
