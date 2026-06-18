package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"path"
	"strings"

	"github.com/minio/minio-go/v7"
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

func (s *MinioFileStore) List(ctx context.Context, prefix string) ([]string, error) {
	prefix = strings.Trim(s.getFullPath(prefix), "/")
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

func (s *MinioFileStore) Read(ctx context.Context, name string) ([]byte, error) {
	fullPath := s.getFullPath(name)
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

func (s *MinioFileStore) WriteFrom(ctx context.Context, name string, r io.Reader) error {
	fullPath := s.getFullPath(name)
	_, err := s.client.PutObject(ctx, s.bucketName, fullPath, r, -1, minio.PutObjectOptions{})
	if err != nil {
		slog.Error("Failed to save object to MinIO", "bucket", s.bucketName, "key", fullPath, "error", err)
	}
	return err
}

func (s *MinioFileStore) Write(ctx context.Context, name string, data []byte) error {
	return s.WriteFrom(ctx, name, bytes.NewReader(data))
}

func (s *MinioFileStore) Remove(ctx context.Context, name string) error {
	fullPath := s.getFullPath(name)
	err := s.client.RemoveObject(ctx, s.bucketName, fullPath, minio.RemoveObjectOptions{})
	if err != nil {
		slog.Error("Failed to delete object from MinIO", "bucket", s.bucketName, "key", fullPath, "error", err)
	}
	return err
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
