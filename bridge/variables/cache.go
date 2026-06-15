package variables

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log/slog"
	"strings"

	"github.com/allegro/bigcache/v3"
	"github.com/sfperusacdev/identitysdk/entities"
)

type variablesCache struct {
	cache *bigcache.BigCache
}

func (c *variablesCache) get(key string, variableName string) *string {
	foundData, err := c.cache.Get(key)
	if err != nil {
		if !errors.Is(err, bigcache.ErrEntryNotFound) {
			slog.Warn("getting variables cache", "key", key, "error", err)
		}
		return nil
	}

	variables, err := decodeVariables(foundData)
	if err != nil {
		slog.Warn("decoding variables cache", "key", key, "error", err)
		return nil
	}

	for _, v := range variables {
		if strings.TrimSpace(v.Name) == strings.TrimSpace(variableName) {
			value := v.Value
			return &value
		}
	}
	return nil
}

func (c *variablesCache) set(key string, variables []entities.Variable) error {
	encodedData, err := encodeVariables(variables)
	if err != nil {
		return err
	}
	return c.cache.Set(key, encodedData)
}

func (c *variablesCache) close() error {
	return c.cache.Close()
}

func encodeVariables(variables []entities.Variable) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(variables); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeVariables(data []byte) ([]entities.Variable, error) {
	var variables []entities.Variable
	if err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(&variables); err != nil {
		return nil, err
	}
	return variables, nil
}
