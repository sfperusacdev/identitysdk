package services

import (
	"context"
	"strings"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/entities"
	variablecache "github.com/sfperusacdev/identitysdk/internal/variable_cache"
)

func (s *ExternalBridgeService) ReadVariableGlobal(ctx context.Context, variableName string) (string, error) {
	var company = "____global____system____domain____"
	_, token := s.readCompanyAndToken(ctx)
	var cachedValue = variablecache.DefaultCache.GetVariable(ctx, company, variableName)
	if cachedValue != nil {
		return strings.TrimSpace(*cachedValue), nil
	}
	var baseUrl = identitysdk.GetIdentityServer()
	var apiresponse struct {
		Message string              `json:"message"`
		Data    []entities.Variable `json:"data"`
	}
	const enpointPath = `/api/v1/global/all/properties`
	if err := s.MakeRequest(ctx,
		baseUrl, enpointPath,
		WithAuthorization(token),
		WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return "", err
	}
	variablecache.DefaultCache.SetVariables(ctx, company, apiresponse.Data)
	for _, v := range apiresponse.Data {
		if strings.TrimSpace(v.Name) == strings.TrimSpace(variableName) {
			return strings.TrimSpace(v.Value), nil
		}
	}
	return "", ErrVariableNotFound
}
