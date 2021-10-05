package runtime

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

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

// check if data is binary content type
func isBinary(data []byte) bool {
	nonBinaryPrefixes := []string{
		"text/",
	}
	commonNonBinaryTypes := map[string]struct{}{
		"application/json":         struct{}{},
		"application/vnd.api+json": struct{}{},
		"image/svg+xml":            struct{}{},
		"application/xhtml+xml":    struct{}{},
		"application/xml":          struct{}{},
	}
	contentType := http.DetectContentType(data)
	for _, p := range nonBinaryPrefixes {
		if strings.HasPrefix(contentType, p) {
			return false
		}
	}
	if _, ok := commonNonBinaryTypes[contentType]; ok {
		return false
	}
	return true
}
