package storage

import "context"

type FileStorer interface {
	Read(ctx context.Context, filepath string) ([]byte, error)
	Save(ctx context.Context, filepath string, data []byte) error
	Delete(ctx context.Context, filepath string) error
	Replace(ctx context.Context, filepath string, data []byte) error
}
