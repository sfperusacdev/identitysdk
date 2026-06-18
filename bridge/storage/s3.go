package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type S3FileStore struct {
	bucketName   string
	subdirectory string
	s3Client     *s3.Client
}

var _ FileStorer = (*S3FileStore)(nil)

func NewS3FileStoreWithClient(ctx context.Context, bucket string, client *s3.Client) (*S3FileStore, error) {
	bucket = strings.Trim(bucket, "/")
	parts := strings.Split(bucket, "/")
	return &S3FileStore{
		bucketName:   parts[0],
		s3Client:     client,
		subdirectory: strings.Join(parts[1:], "/"),
	}, nil
}

func (s *S3FileStore) getFullPath(filepath string) string {
	return path.Join(s.subdirectory, filepath)
}

func (s *S3FileStore) List(ctx context.Context, prefix string) ([]string, error) {
	prefix = s.getFullPath(prefix)
	out, err := s.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(strings.Trim(prefix, "/")),
	})
	if err != nil {
		slog.Error("Failed to list objects from S3", "bucket", s.bucketName, "prefix", prefix, "error", err)
		return nil, err
	}

	keys := make([]string, 0, len(out.Contents))
	for _, item := range out.Contents {
		keys = append(keys, *item.Key)
	}
	return keys, nil
}

func (s *S3FileStore) Read(ctx context.Context, name string) ([]byte, error) {
	fullPath := s.getFullPath(name)
	obj, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) && ae.ErrorCode() == "NoSuchKey" {
			return nil, ErrFileNotFound
		}
		slog.Error("Failed to get object from S3", "bucket", s.bucketName, "key", fullPath, "error", err)
		return nil, err
	}
	defer obj.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, obj.Body); err != nil {
		slog.Error("Failed to read object body from S3", "bucket", s.bucketName, "key", fullPath, "error", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *S3FileStore) WriteFrom(ctx context.Context, name string, r io.Reader) error {
	fullPath := s.getFullPath(name)
	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullPath),
		Body:   r,
	})
	if err != nil {
		slog.Error("Failed to save object to S3", "bucket", s.bucketName, "key", fullPath, "error", err)
	}
	return err
}

func (s *S3FileStore) Write(ctx context.Context, name string, data []byte) error {
	return s.WriteFrom(ctx, name, bytes.NewReader(data))
}

func (s *S3FileStore) Remove(ctx context.Context, name string) error {
	fullPath := s.getFullPath(name)
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		slog.Error("Failed to delete object from S3", "bucket", s.bucketName, "key", fullPath, "error", err)
	}
	return err
}
