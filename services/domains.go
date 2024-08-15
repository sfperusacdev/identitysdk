package services

import (
	"context"

	"github.com/sfperusacdev/identitysdk"
)

func (s *ExternalBridgeService) GetDominios(ctx context.Context) ([]string, error) {
	var apiresponse struct {
		Message string   `json:"message"`
		Data    []string `json:"data"`
	}
	var err = s.makeRequest(ctx,
		identitysdk.GetIdentityServer(),
		"/v1/get-list-empresas", "-", &apiresponse)
	if err != nil {
		return nil, err
	}
	return apiresponse.Data, nil
}
