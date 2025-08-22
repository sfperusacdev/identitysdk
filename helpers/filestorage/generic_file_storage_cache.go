package filestorage

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/sfperusacdev/identitysdk/helpers/filecache"
	"github.com/sfperusacdev/identitysdk/services"
)

type GenericFileStorageWithCache struct {
	storer FileStorage
	cache  *filecache.FileCache
}

var _ FileStorage = (*GenericFileStorageWithCache)(nil)

type CacheOptions struct {
	maxEntries    int
	evictInterval time.Duration
	storer        FileStorage
}

type CacheOption func(*CacheOptions)

func WithMaxEntries(n int) CacheOption {
	return func(opts *CacheOptions) {
		opts.maxEntries = max(10, n)
	}
}

func WithEvictInterval(d time.Duration) CacheOption {
	return func(opts *CacheOptions) {
		opts.evictInterval = max(5*time.Minute, d)
	}
}

func WithCustomFileStorage(st FileStorage) CacheOption {
	return func(opts *CacheOptions) {
		opts.storer = st
	}
}

func NewGenericFileStorageWithCache(
	bridge *services.ExternalBridgeService,
	variableName string,
	baseCacheDir string,
	options ...CacheOption,
) (*GenericFileStorageWithCache, error) {
	cfg := CacheOptions{
		maxEntries:    100,
		evictInterval: time.Hour,
	}
	for _, opt := range options {
		opt(&cfg)
	}
	if cfg.storer == nil {
		cfg.storer = NewGenericFileStorage(variableName, bridge)
	}
	cache, err := filecache.NewFileCache(
		baseCacheDir,
		cfg.maxEntries,
		cfg.evictInterval,
	)
	if err != nil {
		return nil, err
	}

	return &GenericFileStorageWithCache{
		storer: cfg.storer,
		cache:  cache,
	}, nil
}

// Read implements services.FileStorage.
func (t *GenericFileStorageWithCache) Read(ctx context.Context, fileName string) ([]byte, error) {
	cacheContent, err := t.cache.Read(fileName)
	if err == nil {
		return cacheContent, nil
	}
	if !errors.Is(err, filecache.ErrFileNotExists) {
		slog.Warn("failed to read from cache", "error", err, "file", fileName)
	}
	content, err := t.storer.Read(ctx, fileName)
	if err != nil {
		return nil, err
	}
	if err := t.cache.Write(fileName, content); err != nil {
		slog.Warn("failed to write to cache", "error", err, "file", fileName)
	}
	return content, nil
}

// Store implements services.FileStorage.
func (t *GenericFileStorageWithCache) Store(ctx context.Context, fileName string, data []byte) error {
	if err := t.storer.Store(ctx, fileName, data); err != nil {
		return err
	}
	if err := t.cache.Write(fileName, data); err != nil {
		slog.Warn("failed to write to cache", "error", err, "file", fileName)
	}
	return nil
}

// Delete implements services.FileStorage.
func (t *GenericFileStorageWithCache) Delete(ctx context.Context, fileNames []string) error {
	if err := t.storer.Delete(ctx, fileNames); err != nil {
		return err
	}
	// TODO; delete from cache
	return nil
}
