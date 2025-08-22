package filestorage

import "context"

type FileStorage interface {
	Store(ctx context.Context, fileName string, data []byte) error
	Read(ctx context.Context, fileName string) ([]byte, error)
	Delete(ctx context.Context, fileNames []string) error
}
