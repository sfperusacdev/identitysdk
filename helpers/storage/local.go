package storage

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
)

type LocalFileStore struct {
	basePath string
}

var _ FileStorer = (*S3FileStore)(nil)

func NewLocalFileStore(basePath string) *LocalFileStore {
	return &LocalFileStore{basePath: basePath}
}

func (l *LocalFileStore) getFullPath(filePath string) string {
	return filepath.Join(l.basePath, filePath)
}

func (l *LocalFileStore) Read(ctx context.Context, filepath string) ([]byte, error) {
	fullPath := l.getFullPath(filepath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		slog.Error("Failed to read file from local storage",
			"path", fullPath,
			"error", err,
		)
		return nil, err
	}
	return data, nil
}

func (l *LocalFileStore) Save(ctx context.Context, filePath string, data []byte) error {
	fullPath := l.getFullPath(filePath)

	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		slog.Error("Failed to create directories for file",
			"path", fullPath,
			"error", err,
		)
		return err
	}

	err := os.WriteFile(fullPath, data, os.ModePerm)
	if err != nil {
		slog.Error("Failed to save file to local storage",
			"path", fullPath,
			"error", err,
		)
	}
	return err
}

func (l *LocalFileStore) Delete(ctx context.Context, filepath string) error {
	fullPath := l.getFullPath(filepath)
	err := os.Remove(fullPath)
	if err != nil {
		slog.Error("Failed to delete file from local storage",
			"path", fullPath,
			"error", err,
		)
	}
	return err
}

func (l *LocalFileStore) Replace(ctx context.Context, filepath string, data []byte) error {
	if err := l.Delete(ctx, filepath); err != nil && !os.IsNotExist(err) {
		slog.Error("Failed to delete file before replacing",
			"path", filepath,
			"error", err,
		)
		return err
	}

	if err := l.Save(ctx, filepath, data); err != nil {
		slog.Error("Failed to save file while replacing",
			"path", filepath,
			"error", err,
		)
		return err
	}

	return nil
}
