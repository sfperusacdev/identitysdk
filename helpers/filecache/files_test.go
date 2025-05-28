package filecache_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sfperusacdev/identitysdk/helpers/filecache"
)

func TestGetAllFilePaths(t *testing.T) {
	baseDir := t.TempDir()

	os.Mkdir(filepath.Join(baseDir, "sub"), 0755)
	os.Mkdir(filepath.Join(baseDir, "sub", "nested"), 0755)
	os.Mkdir(filepath.Join(baseDir, "another"), 0755)
	os.Mkdir(filepath.Join(baseDir, "another", "deep"), 0755)

	filesToCreate := []string{
		"file1.txt",
		"sub/file2.txt",
		"sub/nested/file3.txt",
		"another/file4.txt",
		"another/deep/file5.txt",
	}

	for _, relPath := range filesToCreate {
		fullPath := filepath.Join(baseDir, relPath)
		os.WriteFile(fullPath, []byte("test"), 0644)
	}

	got, err := filecache.GetAllFilePaths(baseDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := map[string]bool{
		"file1.txt":              true,
		"sub/file2.txt":          true,
		"sub/nested/file3.txt":   true,
		"another/file4.txt":      true,
		"another/deep/file5.txt": true,
	}

	if len(got) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(got))
	}

	for _, path := range got {
		if !expected[path] {
			t.Errorf("Unexpected file path: %s", path)
		}
	}

	os.RemoveAll(baseDir)
}
