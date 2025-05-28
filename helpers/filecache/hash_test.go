package filecache_test

import (
	"path/filepath"
	"testing"

	"github.com/sfperusacdev/identitysdk/helpers/filecache"
)

func TestGetHashedFilePath(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"example.txt", filepath.Join("e1", "bc", "example.txt")},
		{"test.jpg", filepath.Join("cb", "56", "test.jpg")},
		{"data.bin", filepath.Join("1f", "ad", "data.bin")},
		{"report-2025.pdf", filepath.Join("1e", "e0", "report-2025.pdf")},
		{"archive.tar.gz", filepath.Join("cb", "ea", "archive.tar.gz")},
		{"notes.md", filepath.Join("12", "df", "notes.md")},
		{"image_001.png", filepath.Join("c4", "6e", "image_001.png")},
		{"video.mp4", filepath.Join("df", "a3", "video.mp4")},
		{"README", filepath.Join("69", "e2", "README")},
		{"config.yaml", filepath.Join("b5", "10", "config.yaml")},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := filecache.GetHashedFilePath(tt.filename)
			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
