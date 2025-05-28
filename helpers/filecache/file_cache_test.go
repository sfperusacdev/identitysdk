package filecache_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sfperusacdev/identitysdk/helpers/filecache"
)

func TestFileCache_LoadExistingFiles(t *testing.T) {
	baseDir := t.TempDir()

	initialFiles := []struct {
		filename string
		content  []byte
		expected string
	}{
		{"example.txt", []byte("hello world"), filepath.Join(baseDir, "e1", "bc", "example.txt")},
		{"test.jpg", []byte("image content"), filepath.Join(baseDir, "cb", "56", "test.jpg")},
	}

	for _, file := range initialFiles {
		if err := os.MkdirAll(filepath.Dir(file.expected), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(file.expected, file.content, 0644); err != nil {
			t.Fatalf("failed to write initial file: %v", err)
		}
	}

	cache, err := filecache.NewFileCache(baseDir, 10, time.Hour)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	for _, tt := range initialFiles {
		t.Run(tt.filename, func(t *testing.T) {
			data, err := cache.Read(tt.filename)
			if err != nil {
				t.Fatalf("read failed: %v", err)
			}

			if !bytes.Equal(data, tt.content) {
				t.Fatalf("read content mismatch: got %q, want %q", data, tt.content)
			}

			if _, err := os.Stat(tt.expected); err != nil {
				t.Fatalf("expected file does not exist at %q", tt.expected)
			}
		})
	}

	if err := os.RemoveAll(baseDir); err != nil {
		t.Fatalf("failed to clean up temp dir: %v", err)
	}
}

func TestFileCache_WriteAndRead(t *testing.T) {
	baseDir := t.TempDir()

	cache, err := filecache.NewFileCache(baseDir, 10, time.Hour)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	tests := []struct {
		filename string
		content  []byte
		expected string
	}{
		{"example.txt", []byte("hello world"), filepath.Join(baseDir, "e1", "bc", "example.txt")},
		{"test.jpg", []byte("image content"), filepath.Join(baseDir, "cb", "56", "test.jpg")},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			err := cache.Write(tt.filename, tt.content)
			if err != nil {
				t.Fatalf("write failed: %v", err)
			}

			data, err := cache.Read(tt.filename)
			if err != nil {
				t.Fatalf("read failed: %v", err)
			}

			if !bytes.Equal(data, tt.content) {
				t.Fatalf("read content mismatch: got %q, want %q", data, tt.content)
			}

			if _, err := os.Stat(tt.expected); err != nil {
				t.Fatalf("expected file does not exist at %q", tt.expected)
			}
		})
	}
}
