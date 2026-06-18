package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"path"
	"strings"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/panjf2000/ants/v2"
)

type MinioFileStore struct {
	bucketName   string
	subdirectory string
	client       *minio.Client
}

var _ FileStorer = (*MinioFileStore)(nil)

func NewMinioFileStoreWithClient(ctx context.Context, bucket string, client *minio.Client) (*MinioFileStore, error) {
	bucket = strings.Trim(bucket, "/")
	parts := strings.Split(bucket, "/")
	return &MinioFileStore{
		bucketName:   parts[0],
		client:       client,
		subdirectory: strings.Join(parts[1:], "/"),
	}, nil
}

func (s *MinioFileStore) getFullPath(filepath string) string {
	return path.Join(s.subdirectory, filepath)
}

func (s *MinioFileStore) List(ctx context.Context, filepath string) ([]string, error) {
	prefix := strings.Trim(s.getFullPath(filepath), "/")
	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	keys := []string{}
	for object := range objectCh {
		if object.Err != nil {
			slog.Error("Failed to list objects from MinIO", "bucket", s.bucketName, "prefix", prefix, "error", object.Err)
			return nil, object.Err
		}
		keys = append(keys, object.Key)
	}
	return keys, nil
}

func (s *MinioFileStore) Read(ctx context.Context, filepath string) ([]byte, error) {
	fullPath := s.getFullPath(filepath)
	obj, err := s.client.GetObject(ctx, s.bucketName, fullPath, minio.GetObjectOptions{})
	if err != nil {
		if isMinioNotFound(err) {
			return nil, ErrFileNotFound
		}
		slog.Error("Failed to get object from MinIO", "bucket", s.bucketName, "key", fullPath, "error", err)
		return nil, err
	}
	defer obj.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, obj); err != nil {
		if isMinioNotFound(err) {
			return nil, ErrFileNotFound
		}
		slog.Error("Failed to read object body from MinIO", "bucket", s.bucketName, "key", fullPath, "error", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *MinioFileStore) SaveR(ctx context.Context, filePath string, r io.Reader) error {
	fullPath := s.getFullPath(filePath)
	_, err := s.client.PutObject(ctx, s.bucketName, fullPath, r, -1, minio.PutObjectOptions{})
	if err != nil {
		slog.Error("Failed to save object to MinIO", "bucket", s.bucketName, "key", fullPath, "error", err)
	}
	return err
}

func (s *MinioFileStore) Save(ctx context.Context, filePath string, data []byte) error {
	return s.SaveR(ctx, filePath, bytes.NewReader(data))
}

func (s *MinioFileStore) SaveRBatch(ctx context.Context, files map[string]io.Reader) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(files))

	pool, err := ants.NewPool(10)
	if err != nil {
		return err
	}
	defer pool.Release()

	for path, data := range files {
		p := path
		d := data
		wg.Add(1)
		err := pool.Submit(func() {
			defer wg.Done()
			if err := s.SaveR(ctx, p, d); err != nil {
				errCh <- err
			}
		})
		if err != nil {
			wg.Done()
			return err
		}
	}

	wg.Wait()
	close(errCh)
	if len(errCh) > 0 {
		return <-errCh
	}
	return nil
}

func (s *MinioFileStore) SaveBatch(ctx context.Context, files map[string][]byte) error {
	readers := make(map[string]io.Reader, len(files))
	for key, value := range files {
		readers[key] = bytes.NewReader(value)
	}
	return s.SaveRBatch(ctx, readers)
}

func (s *MinioFileStore) Delete(ctx context.Context, filepath string) error {
	fullPath := s.getFullPath(filepath)
	err := s.client.RemoveObject(ctx, s.bucketName, fullPath, minio.RemoveObjectOptions{})
	if err != nil {
		slog.Error("Failed to delete object from MinIO", "bucket", s.bucketName, "key", fullPath, "error", err)
	}
	return err
}

func (s *MinioFileStore) Replace(ctx context.Context, filePath string, data []byte) error {
	if err := s.Delete(ctx, filePath); err != nil {
		slog.Error("Failed to delete object before replacing in MinIO", "bucket", s.bucketName, "key", filePath, "error", err)
		return err
	}
	if err := s.Save(ctx, filePath, data); err != nil {
		slog.Error("Failed to save object while replacing in MinIO", "bucket", s.bucketName, "key", filePath, "error", err)
		return err
	}
	return nil
}

func isMinioNotFound(err error) bool {
	if err == nil {
		return false
	}
	var response minio.ErrorResponse
	if errors.As(err, &response) {
		return response.Code == "NoSuchKey" || response.Code == "NoSuchObject" || response.Code == "NotFound"
	}
	return false
}
