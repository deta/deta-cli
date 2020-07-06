package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
)

const (
	githubRepoRoot = "https://api.github.com/repos/deta/deta-cli"
)

var (
	// set with Makefile during compilation
	detaVersion string
	goVersion   string
	platform    string

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print deta version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(fmt.Sprintf("%s %s %s", rootCmd.Use, detaVersion, platform))
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

type latestRelease struct {
	Tag string `json:"tag_name"`
}

func getLatestVersionTag() (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/releases/latest", githubRepoRoot))
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var lr latestRelease
	err = json.Unmarshal(body, &lr)
	if err != nil {
		return "", err
	}
	return lr.Tag, nil
}

func checkVersionExists(tag string) (bool, error) {
	resp, err := http.Get(fmt.Sprintf("%s/releases/tags/%s", githubRepoRoot, tag))
	if err != nil {
		return false, err
	}
	if resp.StatusCode == 200 {
		return true, nil
	}
	if resp.StatusCode == 404 {
		return false, nil
	}
	return false, fmt.Errorf("unexpected status code from github: %d", resp.StatusCode)
}

func isLowerVersion(version, from string) (bool, error) {
	version = strings.TrimPrefix(version, "v")
	from = strings.TrimPrefix(from, "v")

	va, err := semver.NewVersion(version)
	if err != nil {
		return false, err
	}
	vb, err := semver.NewVersion(from)
	if err != nil {
		return false, err
	}
	return va.LessThan(*vb), nil
}

type checkVersionMsg struct {
	isLower bool
	err     error
}

func checkVersion(c chan *checkVersionMsg) {
	cm := &checkVersionMsg{}
	latestVersion, err := getLatestVersionTag()
	if err != nil {
		cm.err = err
		c <- cm
		return
	}
	lowerVersion, err := isLowerVersion(detaVersion, latestVersion)
	if err != nil {
		cm.err = err
		c <- cm
		return
	}
	cm.isLower = lowerVersion
	cm.err = nil
	c <- cm
	return
}
