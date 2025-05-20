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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/panjf2000/ants/v2"
)

type S3FileStore struct {
	bucketName   string
	subdirectory string
	s3Client     *s3.Client
}

var _ FileStorer = (*S3FileStore)(nil)

func NewS3FileStoreWithClient(ctx context.Context, bucket string, client *s3.Client) (*S3FileStore, error) {
	bucket = strings.Trim(bucket, "/")
	var parts = strings.Split(bucket, "/")
	return &S3FileStore{
		bucketName:   parts[0],
		s3Client:     client,
		subdirectory: strings.Join(parts[1:], "/"),
	}, nil
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

func (s *S3FileStore) List(ctx context.Context, filepath string) ([]string, error) {
	prefix := s.getFullPath(filepath)
	out, err := s.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(strings.Trim(prefix, "/")),
	})
	if err != nil {
		slog.Error("Failed to list objects from S3",
			"bucket", s.bucketName,
			"prefix", prefix,
			"error", err,
		)
		return nil, err
	}

	var keys []string
	for _, item := range out.Contents {
		keys = append(keys, *item.Key)
	}
	return keys, nil
}

func (s *S3FileStore) Read(ctx context.Context, filepath string) ([]byte, error) {
	fullPath := s.getFullPath(filepath)
	obj, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) && ae.ErrorCode() == "NoSuchKey" {
			return nil, ErrFileNotFound
		}
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

func (s *S3FileStore) SaveR(ctx context.Context, path string, r io.Reader) error {
	fullPath := s.getFullPath(path)
	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullPath),
		Body:   r,
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

func (s *S3FileStore) Save(ctx context.Context, path string, data []byte) error {
	return s.SaveR(ctx, path, bytes.NewReader(data))
}

func (s *S3FileStore) SaveRBatch(ctx context.Context, files map[string]io.Reader) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(files))

	pool, err := ants.NewPool(10) // define el número máximo de workers
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

func (s *S3FileStore) SaveBatch(ctx context.Context, files map[string][]byte) error {
	var readers = make(map[string]io.Reader, len(files))
	for key, value := range files {
		readers[key] = bytes.NewBuffer(value)
	}
	return s.SaveRBatch(ctx, readers)
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
