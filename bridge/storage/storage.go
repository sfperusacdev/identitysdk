package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/minio/minio-go/v7"
	miniocredentials "github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sfperusacdev/identitysdk/bridge/variables"
)

const (
	DriverMinio = "minio"
	DriverAWSS3 = "aws-s3"
	DriverLocal = "local"
)

type ReadVariableFunc func(ctx context.Context, key string) (string, error)

type StorageService struct {
	variables *variables.VariablesService
}

func NewStorageService(variables *variables.VariablesService) *StorageService {
	return &StorageService{variables: variables}
}

func (s *StorageService) Create(ctx context.Context, bucket string) (FileStorer, error) {
	return s.CreateWith(ctx, bucket, s.variables.Me.Read)
}

func (s *StorageService) CreateGlobal(ctx context.Context, bucket string) (FileStorer, error) {
	return s.CreateWith(ctx, bucket, s.variables.Global.Read)
}

func (s *StorageService) CreateWith(ctx context.Context, bucket string, readVariable ReadVariableFunc) (FileStorer, error) {
	driver, err := readOptionalVariable(ctx, readVariable, "STORAGE_DRIVER")
	if err != nil {
		return nil, fmt.Errorf("failed to read STORAGE_DRIVER: %w", err)
	}

	switch normalizeDriver(driver) {
	case DriverMinio:
		return s.createMinio(ctx, bucket, readVariable)
	case DriverAWSS3:
		return s.createAWSS3(ctx, bucket, readVariable)
	case DriverLocal:
		return s.createLocal(ctx, bucket, readVariable)
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", driver)
	}
}

func (s *StorageService) createMinio(ctx context.Context, bucket string, readVariable ReadVariableFunc) (FileStorer, error) {
	accessKeyID, err := readRequiredVariableFallback(ctx, readVariable, "STORAGE_MINIO_ACCESS_KEY_ID", "MINIO_ACCESS_KEY_ID")
	if err != nil {
		return nil, err
	}
	secretAccessKey, err := readRequiredVariableFallback(ctx, readVariable, "STORAGE_MINIO_SECRET_ACCESS_KEY", "MINIO_SECRET_ACCESS_KEY")
	if err != nil {
		return nil, err
	}
	endpoint, err := readRequiredVariableFallback(ctx, readVariable, "STORAGE_MINIO_ENDPOINT", "MINIO_ENDPOINT")
	if err != nil {
		return nil, err
	}
	region, err := readOptionalVariableFallback(ctx, readVariable, "STORAGE_MINIO_REGION", "MINIO_REGION")
	if err != nil {
		return nil, fmt.Errorf("failed to read STORAGE_MINIO_REGION: %w", err)
	}
	if region == "" {
		region = "us-east-1"
	}
	if !isValidURL(endpoint) {
		return nil, fmt.Errorf("invalid MINIO_ENDPOINT: %s", endpoint)
	}

	endpointHost, secure, err := parseMinioEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	client, err := minio.New(endpointHost, &minio.Options{
		Creds:  miniocredentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: secure,
		Region: region,
	})
	if err != nil {
		return nil, err
	}
	return NewMinioFileStoreWithClient(ctx, bucket, client)
}

func (s *StorageService) createAWSS3(ctx context.Context, bucket string, readVariable ReadVariableFunc) (FileStorer, error) {
	accessKeyID, err := readOptionalVariableFallback(ctx, readVariable, "STORAGE_AWS_S3_ACCESS_KEY_ID", "AWS_S3_ACCESS_KEY_ID", "AWS_ACCESS_KEY_ID")
	if err != nil {
		return nil, fmt.Errorf("failed to read STORAGE_AWS_S3_ACCESS_KEY_ID: %w", err)
	}
	secretAccessKey, err := readOptionalVariableFallback(ctx, readVariable, "STORAGE_AWS_S3_SECRET_ACCESS_KEY", "AWS_S3_SECRET_ACCESS_KEY", "AWS_SECRET_ACCESS_KEY")
	if err != nil {
		return nil, fmt.Errorf("failed to read STORAGE_AWS_S3_SECRET_ACCESS_KEY: %w", err)
	}
	baseURL, err := readOptionalVariableFallback(ctx, readVariable, "STORAGE_AWS_S3_BASE_URL", "AWS_S3_BASE_URL", "S3_SDK_STORAGE_BASE_URL")
	if err != nil {
		return nil, fmt.Errorf("failed to read STORAGE_AWS_S3_BASE_URL: %w", err)
	}
	region, err := readOptionalVariableFallback(ctx, readVariable, "STORAGE_AWS_S3_REGION", "AWS_S3_REGION", "AWS_REGION")
	if err != nil {
		return nil, fmt.Errorf("failed to read STORAGE_AWS_S3_REGION: %w", err)
	}
	if region == "" {
		region = "us-east-1"
	}

	conf, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	s3Client := s3.NewFromConfig(conf, func(o *s3.Options) {
		if baseURL != "" {
			if !isValidURL(baseURL) {
				slog.Warn("Invalid AWS S3 base URL", "url", baseURL)
			} else {
				o.BaseEndpoint = &baseURL
			}
		}
		if accessKeyID != "" && secretAccessKey != "" {
			o.Credentials = credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")
		}
	})
	return NewS3FileStoreWithClient(ctx, bucket, s3Client)
}

func (s *StorageService) createLocal(ctx context.Context, bucket string, readVariable ReadVariableFunc) (FileStorer, error) {
	basePath, err := readRequiredVariableFallback(ctx, readVariable, "STORAGE_LOCAL_BASE_PATH", "LOCAL_STORAGE_BASE_PATH")
	if err != nil {
		return nil, err
	}
	return NewLocalFileStore(path.Join(basePath, bucket)), nil
}

func normalizeDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", DriverMinio:
		return DriverMinio
	case DriverAWSS3, "s3", "aws":
		return DriverAWSS3
	case DriverLocal:
		return DriverLocal
	default:
		return strings.ToLower(strings.TrimSpace(driver))
	}
}

func readRequiredVariable(ctx context.Context, readVariable ReadVariableFunc, key string) (string, error) {
	value, err := readOptionalVariable(ctx, readVariable, key)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", key, err)
	}
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func readOptionalVariable(ctx context.Context, readVariable ReadVariableFunc, key string) (string, error) {
	value, err := readVariable(ctx, key)
	if err != nil {
		if errors.Is(err, variables.ErrVariableNotFound) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func readRequiredVariableFallback(ctx context.Context, readVariable ReadVariableFunc, primary string, fallbacks ...string) (string, error) {
	value, err := readOptionalVariableFallback(ctx, readVariable, primary, fallbacks...)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", primary, err)
	}
	if value == "" {
		return "", fmt.Errorf("%s is required", primary)
	}
	return value, nil
}

func readOptionalVariableFallback(ctx context.Context, readVariable ReadVariableFunc, primary string, fallbacks ...string) (string, error) {
	keys := append([]string{primary}, fallbacks...)
	for _, key := range keys {
		value, err := readOptionalVariable(ctx, readVariable, key)
		if err != nil || value != "" {
			return value, err
		}
	}
	return "", nil
}

func readBoolVariable(ctx context.Context, readVariable ReadVariableFunc, key string, defaultValue bool) (bool, error) {
	value, err := readOptionalVariable(ctx, readVariable, key)
	if err != nil {
		return false, fmt.Errorf("failed to read %s: %w", key, err)
	}
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid %s: %w", key, err)
	}
	return parsed, nil
}

func isValidURL(rawURL string) bool {
	if !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL
	}
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}
	return parsedURL.Host != ""
}

func parseMinioEndpoint(rawURL string) (string, bool, error) {
	if !strings.Contains(rawURL, "://") {
		return strings.TrimSpace(rawURL), false, nil
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", false, err
	}
	if parsedURL.Host == "" {
		return "", false, fmt.Errorf("invalid MINIO_ENDPOINT: %s", rawURL)
	}
	return parsedURL.Host, parsedURL.Scheme == "https", nil
}
