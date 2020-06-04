package runtime

import "encoding/json"

// DepChanges changes in dependencies
type DepChanges struct {
	Added   []string
	Removed []string
}

// ProgInfo program info
type ProgInfo struct {
	ID      string   `json:"id"`
	Space   int64    `json:"space"`
	Runtime string   `json:"runtime"`
	Name    string   `json:"name"`
	Project string   `json:"project"`
	Deps    []string `json:"deps"`
}

// unmarshals data into a progInfo
func progInfoFromBytes(data []byte) (*ProgInfo, error) {
	var p ProgInfo
	err := json.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
