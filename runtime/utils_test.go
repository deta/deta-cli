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

func TestMain(m *testing.M) {
	exitCode := m.Run()
	teardown()
	os.Exit(exitCode)
}

// this will execute after every test
func teardown() {
	// remove tmp folders
	os.RemoveAll(filepath.Join("testdata", "tmp"))
}

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

func testUnzip(t *testing.T, path, dest string, skipFileNames []string, expectedContent map[string][]byte) {
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open path %s: %v", path, err)
	}

	fileData, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}

	// assert no nil error
	assert.NilError(t, unzip(fileData, dest, skipFileNames), fmt.Sprintf("unzip returned non nil error: %v", err))

	f.Close()

	// seen paths to check later if everything in expected content has been seen
	seenPaths := make(map[string]struct{})

	// check file content
	err = filepath.Walk(dest, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("failed to open file %s: %v", path, err)
		}
		readContent, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatalf("failed to read file %s: %v", path, err)
		}
		f.Close()

		relPath, err := filepath.Rel(dest, path)
		if err != nil {
			t.Fatalf("failed to get rel path for path %s with base %s", path, dest)
		}

		// check if content is ok
		content, ok := expectedContent[relPath]
		assert.Assert(t, ok, fmt.Sprintf("file path not present in expected content for file: %s", path))
		assert.DeepEqual(t, content, readContent)

		// register as seen
		seenPaths[relPath] = struct{}{}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk dir %s: %v", dest, err)
	}

	for p := range expectedContent {
		_, ok := seenPaths[p]
		assert.Assert(t, ok, fmt.Sprintf("file %s not present in the unzipped archive", p))
	}
}

func TestUnzip(t *testing.T) {
	archiveTestDataDir := filepath.Join("testdata", "archives")
	tmpWriteDir := filepath.Join("testdata", "tmp")

	testCases := []struct {
		archiveName     string
		destDir         string
		skipFiles       []string
		expectedContent map[string][]byte
	}{
		{
			"test.zip",
			"test",
			[]string{"test.txt"},
			map[string][]byte{
				"main.py":       []byte(`print("test")`),
				"test/test.txt": []byte("should not be skipped\n"),
			},
		},
	}

	for _, tc := range testCases {
		destDir := filepath.Join(tmpWriteDir, tc.destDir)
		err := os.MkdirAll(destDir, os.ModePerm)
		if err != nil {
			t.Fatalf("failed to create tmp dest dir %s: %v", destDir, err)
		}
		testUnzip(t, filepath.Join(archiveTestDataDir, tc.archiveName), destDir, tc.skipFiles, tc.expectedContent)
	}
}
