package services

import (
	"context"
	"errors"
	"strings"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/entities"
	variablecache "github.com/sfperusacdev/identitysdk/internal/variable_cache"
)

var ErrVariableNotFound = errors.New("variable not found")

func (s *ExternalBridgeService) ReadVariable(ctx context.Context, variableName string) (string, error) {
	var company, token = s.readCompanyAndToken(ctx)
	var cachedValue = variablecache.DefaultCache.GetVariable(ctx, company, variableName)
	if cachedValue != nil {
		return strings.TrimSpace(*cachedValue), nil
	}
	var baseUrl = identitysdk.GetIdentityServer()
	const enpointPath = `/api/v1/companies/properties`
	var apiresponse struct {
		Message string              `json:"message"`
		Data    []entities.Variable `json:"data"`
	}
	if err := s.makeRequest(ctx, baseUrl, enpointPath, token, &apiresponse); err != nil {
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
