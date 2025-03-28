package storage

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/url"
	"os"
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

var _ FileStorer = (*S3FileStore)(nil)

func isValidURL(rawURL string) bool {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}
	if parsedURL.Host == "" {
		return false
	}
	return true
}

const envS3BaseURL = "S3_SDK_STORAGE_BASE_URL"

func NewS3FileStore(ctx context.Context, awsRegion, bucketBasePath string) (*S3FileStore, error) {
	conf, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		return nil, err
	}
	var parts = strings.Split(bucketBasePath, "/")
	return &S3FileStore{
		bucketName: parts[0],
		s3Client: s3.NewFromConfig(conf, func(o *s3.Options) {
			value, exists := os.LookupEnv(envS3BaseURL)
			if !exists {
				return
			}
			if !isValidURL(value) {
				slog.Warn("Invalid S3 base URL", "variable", envS3BaseURL, "url", value)
				return
			}
			o.BaseEndpoint = &value
			slog.Info("S3 base URL successfully configured", "url", value)
		}),
		subdirectory: strings.Join(parts[1:], "/"),
	}, nil
}

func NewS3FileStoreWithBaseURL(ctx context.Context, baseURL, awsRegion, bucketBasePath string) (*S3FileStore, error) {
	conf, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		return nil, err
	}

	var parts = strings.Split(bucketBasePath, "/")

	s3Client := s3.NewFromConfig(conf, func(o *s3.Options) {
		if baseURL == "" {
			slog.Warn("S3 base URL is empty, using default AWS endpoint")
			return
		}
		if !isValidURL(baseURL) {
			slog.Warn("Invalid S3 base URL", "url", baseURL)
			return
		}
		o.BaseEndpoint = &baseURL
		slog.Info("S3 base URL successfully configured", "url", baseURL)
	})

	return &S3FileStore{
		bucketName:   parts[0],
		s3Client:     s3Client,
		subdirectory: strings.Join(parts[1:], "/"),
	}, nil
}

func (s *S3FileStore) getFullPath(filepath string) string {
	return path.Join(s.subdirectory, filepath)
}

func (s *S3FileStore) Read(ctx context.Context, filepath string) ([]byte, error) {
	fullPath := s.getFullPath(filepath)
	obj, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		slog.Error("Failed to get object from S3",
			"bucket", s.bucketName,
			"key", fullPath,
			"error", err,
		)
		return nil, err
	}
	defer obj.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, obj.Body); err != nil {
		slog.Error("Failed to read object body from S3",
			"bucket", s.bucketName,
			"key", fullPath,
			"error", err,
		)
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *S3FileStore) Save(ctx context.Context, path string, data []byte) error {
	fullPath := s.getFullPath(path)
	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullPath),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		slog.Error("Failed to save object to S3",
			"bucket", s.bucketName,
			"key", fullPath,
			"error", err,
		)
	}
	return err
}

func (s *S3FileStore) Delete(ctx context.Context, filepath string) error {
	fullPath := s.getFullPath(filepath)
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		slog.Error("Failed to delete object from S3",
			"bucket", s.bucketName,
			"key", fullPath,
			"error", err,
		)
	}
	return err
}

func (s *S3FileStore) Replace(ctx context.Context, path string, data []byte) error {
	if err := s.Delete(ctx, path); err != nil {
		slog.Error("Failed to delete object before replacing in S3",
			"bucket", s.bucketName,
			"key", path,
			"error", err,
		)
		return err
	}

	if err := s.Save(ctx, path, data); err != nil {
		slog.Error("Failed to save object while replacing in S3",
			"bucket", s.bucketName,
			"key", path,
			"error", err,
		)
		return err
	}

	return nil
}
