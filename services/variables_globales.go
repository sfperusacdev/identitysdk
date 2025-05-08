package services

import (
	"context"
	"strings"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/entities"
	variablecache "github.com/sfperusacdev/identitysdk/internal/variable_cache"
	"github.com/sfperusacdev/identitysdk/xreq"
)

func (s *ExternalBridgeService) ReadVariableGlobal(ctx context.Context, variableName string) (string, error) {
	var company = "____global____system____domain____"
	var cachedValue = variablecache.DefaultCache.GetVariable(ctx, company, variableName)
	if cachedValue != nil {
		return strings.TrimSpace(*cachedValue), nil
	}
	var baseUrl = identitysdk.GetIdentityServer()
	var apiresponse struct {
		Message string              `json:"message"`
		Data    []entities.Variable `json:"data"`
	}
	const enpointPath = `/api/v1/public/global/all/properties`
	if err := xreq.MakeRequest(ctx,
		baseUrl, enpointPath,
		xreq.WithAccessToken(identitysdk.GetAccessToken()),
		xreq.WithUnmarshalResponseInto(&apiresponse),
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
