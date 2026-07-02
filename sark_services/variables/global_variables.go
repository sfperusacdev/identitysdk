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

const globalVariablesCompany = "____global____system____domain____"

type GlobalVariablesService struct {
	cache    *variablesCache
	identity bridgeidentity.IdentityProvider
}

func NewGlobalVariablesService(
	identity bridgeidentity.IdentityProvider,
) (*GlobalVariablesService, error) {
	cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(time.Minute))
	if err != nil {
		return nil, fmt.Errorf("creating global variables cache: %w", err)
	}

	s := &GlobalVariablesService{
		cache:    &variablesCache{cache: cache},
		identity: identity,
	}
	return s, nil
}

func (s *GlobalVariablesService) Read(ctx context.Context, variableName string) (string, error) {
	var cachedValue = s.cache.get(globalVariablesCompany, variableName)
	if cachedValue != nil {
		return strings.TrimSpace(*cachedValue), nil
	}

	var baseUrl = s.identity.IdentityServer()
	var apiresponse struct {
		Message string              `json:"message"`
		Data    []entities.Variable `json:"data"`
	}
	const enpointPath = `/api/v1/public/global/all/properties`
	if err := xreq.MakeRequest(ctx,
		baseUrl, enpointPath,
		xreq.WithAccessToken(s.identity.AccessToken()),
		xreq.WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return "", err
	}
	if err := s.cache.set(globalVariablesCompany, apiresponse.Data); err != nil {
		slog.Warn("setting global variables cache", "error", err)
	}
	for _, v := range apiresponse.Data {
		if strings.TrimSpace(v.Name) == strings.TrimSpace(variableName) {
			return strings.TrimSpace(v.Value), nil
		}
	}
	return "", ErrVariableNotFound
}
