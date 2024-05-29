package internal

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/sfperusacdev/identitysdk/entities"
)

var maxchache = 10 * time.Second

type MemCacheService interface {
	Set(ctx context.Context, token string, data entities.JwtData)
	Get(ctx context.Context, token string) *entities.JwtData
}

type sessioncache struct {
	cache *bigcache.BigCache
	err   error
}
type jwtData struct {
	ID  string `json:"id"`
	Exp int64  `json:"exp"`
}

func (*sessioncache) tokenData(token string) (jwtData, error) {
	var parts = strings.Split(token, ".")
	if len(parts) != 3 {
		return jwtData{}, errors.New("invalid jwt token value")
	}
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return jwtData{}, err
	}
	var jsonObject jwtData
	if err := json.Unmarshal(data, &jsonObject); err != nil {
		return jwtData{}, err
	}
	if jsonObject.ID == "" {
		return jwtData{}, errors.New("invalid jwt toke id")
	}
	return jsonObject, nil
}

func (s *sessioncache) Validar(ctx context.Context, token string) (jwtData, error) {
	info, err := s.tokenData(token)
	if err != nil {
		slog.Error("processing token", "error", err, "token", token)
		return jwtData{}, err
	}
	if s.err != nil {
		slog.Info("Cache error detected: discarding value",
			"cacheKey", info.ID,
			"action", "discard")
		return jwtData{}, err
	}
	var exp = time.Unix(info.Exp, 0)
	var now = time.Now().Add(-1 * maxchache)
	if now.Before(exp) {
		return jwtData{}, err
	}
	return info, nil
}

func (*sessioncache) encodeGob(data entities.JwtData) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (*sessioncache) decodeGob(data []byte) (entities.JwtData, error) {
	var jwtdata entities.JwtData
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&jwtdata)
	if err != nil {
		return entities.JwtData{}, err
	}
	return jwtdata, nil
}

func (s *sessioncache) Set(ctx context.Context, token string, data entities.JwtData) {
	info, err := s.Validar(ctx, token)
	if err != nil {
		return
	}
	encodedData, err := s.encodeGob(data)
	if err != nil {
		slog.Error("Failed to encode data using Gob", "cacheKey", info.ID, "error", err)
		return
	}
	if err := s.cache.Set(info.ID, encodedData); err != nil {
		slog.Error("Failed to set data in cache",
			"cacheKey", info.ID,
			"error", err.Error(),
			"action", "cache_set")
	}
}

func (s *sessioncache) Get(ctx context.Context, token string) *entities.JwtData {
	info, err := s.Validar(ctx, token)
	if err != nil {
		return nil
	}
	foundData, err := s.cache.Get(info.ID)
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			return nil
		}
		slog.Error("Failed to retrieve data from cache", "cacheKey", info.ID, "action", "discard")
	}
	jwtData, err := s.decodeGob(foundData)
	if err != nil {
		slog.Error("Failed to decode data using Gob", "cacheKey", info.ID, "error", err)
		return nil
	}
	return &jwtData
}

var DefaultCache MemCacheService

func init() {
	cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(time.Minute))
	if err != nil {
		slog.Warn("creating cache")
	}
	DefaultCache = &sessioncache{cache, err}
}
