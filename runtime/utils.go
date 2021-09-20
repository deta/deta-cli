package runtime

import (
	"bufio"
	"bytes"
)

// other binary extensions
var otherBinaryExts = map[string]struct{}{
	".mo": {},
}

func readLines(data []byte) ([]string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// contains checks if the given string exists on given array
func contains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}

	return false
}
