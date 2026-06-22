package storage

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type LocalFileStore struct {
	basePath string
}

var _ FileStorer = (*LocalFileStore)(nil)

func NewLocalFileStore(basePath string) *LocalFileStore {
	return &LocalFileStore{basePath: basePath}
}

func (l *LocalFileStore) getFullPath(filePath string) string {
	return filepath.Join(l.basePath, filePath)
}

func (l *LocalFileStore) List(ctx context.Context, prefix string) ([]string, error) {
	fullPath := l.getFullPath(prefix)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		slog.Error("Failed to list directory from local storage", "path", fullPath, "error", err)
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

func (l *LocalFileStore) Read(ctx context.Context, name string) ([]byte, error) {
	fullPath := l.getFullPath(name)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrFileNotFound
		}
		slog.Error("Failed to read file from local storage", "path", fullPath, "error", err)
		return nil, err
	}
	return data, nil
}

func (l *LocalFileStore) Write(ctx context.Context, name string, data []byte) error {
	fullPath := l.getFullPath(name)
	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		slog.Error("Failed to create directories for file", "path", fullPath, "error", err)
		return err
	}

	if err := os.WriteFile(fullPath, data, os.ModePerm); err != nil {
		slog.Error("Failed to save file to local storage", "path", fullPath, "error", err)
		return err
	}
	return nil
}

func (l *LocalFileStore) WriteFrom(ctx context.Context, name string, r io.Reader) error {
	fileBytes, err := io.ReadAll(r)
	if err != nil {
		slog.Error("Error reading file data from the provided reader", "error", err)
		return err
	}
	return l.Write(ctx, name, fileBytes)
}

func (l *LocalFileStore) Remove(ctx context.Context, name string) error {
	fullPath := l.getFullPath(name)
	if err := os.Remove(fullPath); err != nil {
		slog.Error("Failed to delete file from local storage", "path", fullPath, "error", err)
		return err
	}
	return nil
}
