package logic

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	detaVersion string
	platform    string
)

const (
	githubRepoRoot = "https://api.github.com/repos/deta/deta-cli"
)

type LatestRelease struct {
	Tag        string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
}

func getLatestVersion() (*LatestRelease, error) {
	resp, err := http.Get(fmt.Sprintf("%s/releases/latest", githubRepoRoot))
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	lr := &LatestRelease{}
	err = json.Unmarshal(body, lr)
	if err != nil {
		return nil, err
	}
	return lr, nil
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

type checkVersionMsg struct {
	isLower bool
	err     error
}

func checkVersion(c chan *checkVersionMsg) {
	cm := &checkVersionMsg{}
	latestVersion, err := getLatestVersion()
	if err != nil {
		fmt.Println("error in get latest version tag: ", err)
		cm.err = err
		c <- cm
		return
	}
	cm.isLower = detaVersion != latestVersion.Tag && !latestVersion.Prerelease
	cm.err = nil
	c <- cm
}
