package runtime

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
)

func testIsBinary(t *testing.T, testDataDir string, shouldBeBinary bool) {
	err := filepath.Walk(testDataDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			t.Fatalf("failed to access path %v: %v", path, err)
		}
		if info.IsDir() {
			// do nothing for dirs
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			t.Fatalf("failed to open file %v: %v", path, err)
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			t.Fatalf("failed to read file %v: %v", path, err)
		}
		isFileBinary := isBinary(data)
		assert.Equal(t, isFileBinary, shouldBeBinary, fmt.Sprintf("for file %s", path))
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk dir %v: %v", testDataDir, err)
	}
}

func TestIsBinary(t *testing.T) {
	binTestDataDir := filepath.Join("testdata", "binary")
	nonBinTestDataDir := filepath.Join("testdata", "non_binary")
	testIsBinary(t, binTestDataDir, true)
	testIsBinary(t, nonBinTestDataDir, false)
}
