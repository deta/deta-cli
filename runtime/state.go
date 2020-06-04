package runtime

import "encoding/json"

// map filepath to checksum
type stateMap map[string]string

// unmarshals data into a stateMap
func stateMapFromBytes(data []byte) (stateMap, error) {
	var s stateMap
	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// StateChanges changes in state of files of the root directory
type StateChanges struct {
	Changes   map[string][]byte // map of files to content
	Deletions []string
}
