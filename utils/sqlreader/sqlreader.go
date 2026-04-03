package sqlreader

import (
	"io/fs"
	"os"
	"path"
	"strings"
)

type SQLFile struct {
	Path    string
	Name    string
	Content string
}

func LoadSQLFiles(fsys fs.FS, searchRoot string) ([]SQLFile, error) {
	var sqlFiles []SQLFile

	err := fs.WalkDir(fsys, searchRoot, func(currentPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			return nil
		}

		if !strings.EqualFold(path.Ext(entry.Name()), ".sql") {
			return nil
		}

		content, err := fs.ReadFile(fsys, currentPath)
		if err != nil {
			return err
		}

		sqlFiles = append(sqlFiles, SQLFile{
			Path:    currentPath,
			Name:    entry.Name(),
			Content: string(content),
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return sqlFiles, nil
}

func LoadSQLFilesFromPath(dirPath string) ([]SQLFile, error) {
	return LoadSQLFiles(os.DirFS(dirPath), ".")
}
