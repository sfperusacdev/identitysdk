package storage

import (
	"context"
	"errors"
	"io"
)

var ErrFileNotFound = errors.New("file not found")

type FileStorer interface {
	Read(ctx context.Context, filepath string) ([]byte, error)
	Save(ctx context.Context, filepath string, data []byte) error
	SaveR(ctx context.Context, filepath string, r io.Reader) error
	Delete(ctx context.Context, filepath string) error
	Replace(ctx context.Context, filepath string, data []byte) error
	List(ctx context.Context, filepath string) ([]string, error)
	SaveBatch(ctx context.Context, files map[string][]byte) error
	SaveRBatch(ctx context.Context, files map[string]io.Reader) error
}
