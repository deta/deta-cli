package runtime

import "encoding/json"

// DepChanges changes in dependencies
type DepChanges struct {
	Added   []string
	Removed []string
}

// EnvChanges changes in env vars keys
type EnvChanges struct {
	Added   map[string]string
	Removed []string
}

// ProgInfo program info
type ProgInfo struct {
	ID      string   `json:"id"`
	Space   int64    `json:"space"`
	Runtime string   `json:"runtime"`
	Name    string   `json:"name"`
	Path    string   `json:"path"`
	Project string   `json:"project"`
	Account string   `json:"account"`
	Region  string   `json:"region"`
	Deps    []string `json:"deps"`
	Envs    []string `json:"envs"`
	Public  bool     `json:"public"`
}

// unmarshals data into a ProgInfo
func progInfoFromBytes(data []byte) (*ProgInfo, error) {
	var p ProgInfo
	err := json.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// UserInfo user info
type UserInfo struct {
	DefaultSpace   int64  `json:"default_space"`
	DefaultProject string `json:"default_project"`
}

// unmarshals data into a UserInfo
func userInfoFromBytes(data []byte) (*UserInfo, error) {
	var u UserInfo
	err := json.Unmarshal(data, &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
