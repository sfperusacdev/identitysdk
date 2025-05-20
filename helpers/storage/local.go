package storage

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
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

func (l *LocalFileStore) List(ctx context.Context, filepath string) ([]string, error) {
	fullPath := l.getFullPath(filepath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		slog.Error("Failed to list directory from local storage",
			"path", fullPath,
			"error", err,
		)
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

func (l *LocalFileStore) Read(ctx context.Context, filepath string) ([]byte, error) {
	fullPath := l.getFullPath(filepath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrFileNotFound
		}
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

func (l *LocalFileStore) SaveR(ctx context.Context, filePath string, r io.Reader) error {
	fileBytes, err := io.ReadAll(r)
	if err != nil {
		slog.Error("Error reading file data from the provided reader",
			"error", err,
		)
	}
	return l.Save(ctx, filePath, fileBytes)
}

func (l *LocalFileStore) SaveBatch(ctx context.Context, files map[string][]byte) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(files))

	for path, data := range files {
		p := path
		d := data

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := l.Save(ctx, p, d); err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		return <-errCh
	}
	return nil
}

func (l *LocalFileStore) SaveRBatch(ctx context.Context, files map[string]io.Reader) error {
	filesBytes := make(map[string][]byte, len(files))
	for name, file := range files {
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			slog.Error("Error reading file data for batch save",
				"file", name,
				"error", err,
			)
			return err
		}
		filesBytes[name] = fileBytes
	}
	return l.SaveBatch(ctx, filesBytes)
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
