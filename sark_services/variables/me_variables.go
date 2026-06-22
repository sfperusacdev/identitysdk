package variables

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/sfperusacdev/identitysdk/entities"
	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
	"github.com/sfperusacdev/identitysdk/xreq"
)

const meVariablesEndpoint = `/api/v1/companies/me/properties`

type MeVariablesService struct {
	cache    *variablesCache
	identity bridgeidentity.Provider
}

func NewMeVariablesService(
	identity bridgeidentity.Provider,
) (*MeVariablesService, error) {
	cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(time.Minute))
	if err != nil {
		return nil, fmt.Errorf("creating me variables cache: %w", err)
	}

	s := &MeVariablesService{
		cache:    &variablesCache{cache: cache},
		identity: identity,
	}
	return s, nil
}

func (s *MeVariablesService) Read(ctx context.Context, variableName string) (string, error) {
	token := s.identity.SessionToken(ctx)
	cachedValue := s.cache.get(token, variableName)
	if cachedValue != nil {
		return strings.TrimSpace(*cachedValue), nil
	}

	baseUrl := s.identity.IdentityServer()
	apiresponse := struct {
		Message string              `json:"message"`
		Data    []entities.Variable `json:"data"`
	}{}
	if err := xreq.MakeRequest(ctx,
		baseUrl, meVariablesEndpoint,
		xreq.WithAuthorization(token),
		xreq.WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return "", err
	}
	if err := s.cache.set(token, apiresponse.Data); err != nil {
		slog.Warn("setting me variables cache", "error", err)
	}
	for _, v := range apiresponse.Data {
		if strings.TrimSpace(v.Name) == strings.TrimSpace(variableName) {
			return strings.TrimSpace(v.Value), nil
		}
	}
	return "", ErrVariableNotFound
}
