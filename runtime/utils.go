package runtime

import (
	"bufio"
	"bytes"
)

func readLines(data []byte) []string {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}
