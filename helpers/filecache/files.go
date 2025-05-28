package filecache

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func GetAllFilePaths(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			filePath := strings.TrimPrefix(path, root)
			paths = append(paths, strings.TrimLeft(filePath, "/"))
		}
		return nil
	})
	return paths, err
}
