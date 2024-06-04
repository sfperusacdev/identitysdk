package variablecache

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/sfperusacdev/identitysdk/entities"
)

type MemVariablesCacheService interface {
	SetVariables(ctx context.Context, empresa string, variables []entities.Variable)
	GetVariable(ctx context.Context, empresa string, variableName string) *string
}

type sessioncache struct {
	cache *bigcache.BigCache
	err   error
}

func (*sessioncache) encodeGob(variables []entities.Variable) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(variables)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (*sessioncache) decodeGob(data []byte) ([]entities.Variable, error) {
	var variables []entities.Variable
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&variables)
	if err != nil {
		return nil, err
	}
	return variables, nil
}

func (s *sessioncache) SetVariables(ctx context.Context, empresa string, variables []entities.Variable) {
	encodedData, err := s.encodeGob(variables)
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

func (s *sessioncache) GetVariable(ctx context.Context, empresa string, variableName string) *string {
	foundData, err := s.cache.Get(empresa)
	if err != nil {
		if !errors.Is(err, bigcache.ErrEntryNotFound) {
			slog.Error("Failed to retrieve data from cache", "cacheKey", empresa, "action", "discard")
		}
		return nil
	}
	variables, err := s.decodeGob(foundData)
	if err != nil {
		slog.Error("Failed to decode data using Gob", "cacheKey", empresa, "error", err)
		return nil
	}
	for _, v := range variables {
		if strings.TrimSpace(v.Name) == strings.TrimSpace(variableName) {
			var value = v.Value
			return &value
		}
	}
	return nil
}

var DefaultCache MemVariablesCacheService

func init() {
	cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(time.Minute))
	if err != nil {
		slog.Warn("creating cache")
	}
	DefaultCache = &sessioncache{cache, err}
}
