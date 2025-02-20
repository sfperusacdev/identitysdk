package helpers

import (
	"bytes"
	"context"
	"io"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3FileStore struct {
	bucketName   string
	subdirectory string
	s3Client     *s3.Client
}

func NewS3FileStore(ctx context.Context, awsRegion, bucketBasePath string) (*S3FileStore, error) {
	conf, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		return nil, err
	}
	var parts = strings.Split(bucketBasePath, "/")
	return &S3FileStore{
		bucketName:   parts[0],
		s3Client:     s3.NewFromConfig(conf),
		subdirectory: strings.Join(parts[1:], "/"),
	}, nil
}

func (s *S3FileStore) getFullPath(filepath string) string {
	return path.Join(s.subdirectory, filepath)
}

func (s *S3FileStore) Read(ctx context.Context, filepath string) []byte {
	fullPath := s.getFullPath(filepath)
	obj, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		return nil
	}
	defer obj.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, obj.Body); err != nil {
		return nil
	}

	return buf.Bytes()
}

func (s *S3FileStore) Save(ctx context.Context, path string, data []byte) error {
	fullPath := s.getFullPath(path)
	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullPath),
		Body:   bytes.NewReader(data),
	})
	return err
}

func (s *S3FileStore) Delete(ctx context.Context, filepath string) error {
	fullPath := s.getFullPath(filepath)
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullPath),
	})
	return err
}

func (s *S3FileStore) Replace(ctx context.Context, path string, data []byte) error {
	if err := s.Delete(ctx, path); err != nil {
		return err
	}
	return s.Save(ctx, path, data)
}
