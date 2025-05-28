package filecache

import (
	"path/filepath"
	"testing"
)

func Test_exists(t *testing.T) {
	c := &FileCache{
		baseDir:  "/tmp/cache",
		pathsMap: make(map[string]struct{}),
	}
	filename := "doc.pdf"
	hashed := GetHashedFilePath(filename)
	c.pathsMap[hashed] = struct{}{}

	path, ok := c.exists(filename)
	expected := filepath.Join("/tmp/cache", hashed)

	if !ok {
		t.Fatalf("expected file to exist in cache")
	}
	if path != expected {
		t.Fatalf("expected path %s, got %s", expected, path)
	}
}

func Test_addToCache(t *testing.T) {
	c := &FileCache{
		pathsMap: make(map[string]struct{}),
	}
	filename := "new.pdf"
	hashed := GetHashedFilePath(filename)

	c.addToCache(filename)

	if _, ok := c.pathsMap[hashed]; !ok {
		t.Fatalf("expected %s to be added to cache", hashed)
	}
}

func Test_removeFromCache(t *testing.T) {
	filename := "delete.pdf"
	hashed := GetHashedFilePath(filename)
	c := &FileCache{
		pathsMap: map[string]struct{}{hashed: {}},
	}

	c.removeFromCache(filename)

	if _, ok := c.pathsMap[hashed]; ok {
		t.Fatalf("expected %s to be removed from cache", hashed)
	}
}
