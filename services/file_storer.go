package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sfperusacdev/identitysdk/helpers/storage"
)

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

func (s *ExternalBridgeService) CreateNewFileStorer(ctx context.Context, bucket string) (storage.FileStorer, error) {
	accessKeyID, err := s.ReadVariable(ctx, "AWS_ACCESS_KEY_ID")
	if err != nil && !errors.Is(err, ErrVariableNotFound) {
		return nil, fmt.Errorf("failed to read AWS_ACCESS_KEY_ID: %w", err)
	}

	secretAccessKey, err := s.ReadVariable(ctx, "AWS_SECRET_ACCESS_KEY")
	if err != nil && !errors.Is(err, ErrVariableNotFound) {
		return nil, fmt.Errorf("failed to read AWS_SECRET_ACCESS_KEY: %w", err)
	}
	baseURL, err := s.ReadVariable(ctx, "S3_SDK_STORAGE_BASE_URL")
	if err != nil && !errors.Is(err, ErrVariableNotFound) {
		return nil, fmt.Errorf("failed to read S3_SDK_STORAGE_BASE_URL: %w", err)
	}

	region, err := s.ReadVariable(ctx, "AWS_REGION")
	if err != nil && !errors.Is(err, ErrVariableNotFound) {
		return nil, fmt.Errorf("failed to read AWS_REGION: %w", err)
	}
	baseURL = strings.TrimSpace(baseURL)
	accessKeyID = strings.TrimSpace(accessKeyID)
	secretAccessKey = strings.TrimSpace(secretAccessKey)
	region = strings.TrimSpace(region)

	if region == "" {
		region = "us-east-1"
	}

	conf, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

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
		if accessKeyID != "" && secretAccessKey != "" {
			o.Credentials = credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")
		}
		slog.Info("S3 base URL successfully configured", "url", baseURL)
	})
	return storage.NewS3FileStoreWithClient(ctx, bucket, s3Client)
}
