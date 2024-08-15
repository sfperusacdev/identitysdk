package integracioncache

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"log/slog"
	"time"

	"github.com/allegro/bigcache/v3"
)

type IntegracionState struct {
	ExternalReff   string `json:"external_reff"`
	IntegrationURL string `json:"integration_url"`
}

type MemCacheService interface {
	Set(ctx context.Context, empresa string, state IntegracionState)
	Get(ctx context.Context, empresa string) *IntegracionState
}

type sessioncache struct {
	cache *bigcache.BigCache
	err   error
}

func (*sessioncache) encodeGob(value IntegracionState) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(value)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (*sessioncache) decodeGob(data []byte) (*IntegracionState, error) {
	var record *IntegracionState
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (s *sessioncache) Set(ctx context.Context, empresa string, record IntegracionState) {
	encodedData, err := s.encodeGob(record)
	if err != nil {
		slog.Error("Failed to encode data using Gob", "cacheKey", empresa, "error", err)
		return
	}
	if err := s.cache.Set(empresa, encodedData); err != nil {
		slog.Error("Failed to set data in cache",
			"cacheKey", empresa,
			"error", err.Error(),
			"action", "cache_set")
	}
}

func (s *sessioncache) Get(ctx context.Context, empresa string) *IntegracionState {
	foundData, err := s.cache.Get(empresa)
	if err != nil {
		if !errors.Is(err, bigcache.ErrEntryNotFound) {
			slog.Error("Failed to retrieve data from cache", "cacheKey", empresa, "action", "discard")
		}
		return nil
	}
	record, err := s.decodeGob(foundData)
	if err != nil {
		slog.Error("Failed to decode data using Gob", "cacheKey", empresa, "error", err)
		return nil
	}
	return record
}

var DefaultCache MemCacheService

func init() {
	cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(time.Minute))
	if err != nil {
		slog.Warn("creating cache")
	}
	DefaultCache = &sessioncache{cache, err}
}
