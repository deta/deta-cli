package cmd

import (
	"encoding/json"
)

func prettyPrint(data interface{}) (string, error) {
	marshalled, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return "", err
	}
	return string(marshalled), nil
}
