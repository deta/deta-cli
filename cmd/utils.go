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

func removeFromSlice(slice []string, toRemove string) []string {
	for i, s := range slice {
		if s == toRemove {
			slice[i] = slice[len(slice)-1]
			return slice[:len(slice)-1]
		}
	}
	return slice
}

func inSlice(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
