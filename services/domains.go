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
	if err := s.MakeRequest(ctx,
		identitysdk.GetIdentityServer(),
		"/v1/get-list-empresas",
		WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return nil, err
	}
	return apiresponse.Data, nil
}

type Empresa struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (s *ExternalBridgeService) GetEmpresas(ctx context.Context) ([]Empresa, error) {
	var apiresponse struct {
		Message string    `json:"message"`
		Data    []Empresa `json:"data"`
	}
	if err := s.MakeRequest(ctx,
		identitysdk.GetIdentityServer(),
		"/v1/get-list-empresas-full",
		WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return nil, err
	}
	return apiresponse.Data, nil
}
