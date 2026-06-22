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

type VariablesService struct {
	Global   *GlobalVariablesService
	Me       *MeVariablesService
	cache    *variablesCache
	identity bridgeidentity.Provider
}

func NewVariablesService(
	global *GlobalVariablesService,
	me *MeVariablesService,
	identity bridgeidentity.Provider,
) (*VariablesService, error) {
	cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(time.Minute))
	if err != nil {
		return nil, fmt.Errorf("creating variables cache: %w", err)
	}

	s := &VariablesService{
		Global:   global,
		Me:       me,
		cache:    &variablesCache{cache: cache},
		identity: identity,
	}
	return s, nil
}

func (s *VariablesService) Read(ctx context.Context, company string, variableName string) (string, error) {
	var cachedValue = s.cache.get(company, variableName)
	if cachedValue != nil {
		return strings.TrimSpace(*cachedValue), nil
	}

	var baseUrl = s.identity.IdentityServer()
	var enpointPath = fmt.Sprintf("/api/v1/public/companies/%s/properties", company)
	var apiresponse struct {
		Message string              `json:"message"`
		Data    []entities.Variable `json:"data"`
	}
	if err := xreq.MakeRequest(ctx,
		baseUrl, enpointPath,
		xreq.WithAccessToken(s.identity.AccessToken()),
		xreq.WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return "", err
	}
	if err := s.cache.set(company, apiresponse.Data); err != nil {
		slog.Warn("setting variables cache", "company", company, "error", err)
	}
	for _, v := range apiresponse.Data {
		if strings.TrimSpace(v.Name) == strings.TrimSpace(variableName) {
			return strings.TrimSpace(v.Value), nil
		}
	}
	return "", ErrVariableNotFound
}
