package runtime

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
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

// checks if dir is empty
func isDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	names, err := f.Readdirnames(-1)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return true, nil
		}
		return false, err
	}
	return len(names) == 0, nil
}
