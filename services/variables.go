package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/entities"
	variablecache "github.com/sfperusacdev/identitysdk/internal/variable_cache"
)

var ErrVariableNotFound = errors.New("variable not found")

func (s *ExternalBridgeService) ReadVariable(ctx context.Context, variableName string) (string, error) {
	var company, _ = s.readCompanyAndToken(ctx)
	return s.ReadCompanyVariable(ctx, company, variableName)
}

func (s *ExternalBridgeService) ReadCompanyVariable(ctx context.Context, company string, variableName string) (string, error) {
	var cachedValue = variablecache.DefaultCache.GetVariable(ctx, company, variableName)
	if cachedValue != nil {
		return strings.TrimSpace(*cachedValue), nil
	}

	var baseUrl = identitysdk.GetIdentityServer()
	var enpointPath = fmt.Sprintf("/api/v1/public/companies/%s/properties", company)
	var apiresponse struct {
		Message string              `json:"message"`
		Data    []entities.Variable `json:"data"`
	}
	if err := s.MakeRequest(ctx,
		baseUrl, enpointPath,
		WithHeader("X-Access-Token", identitysdk.GetAccessToken()),
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
