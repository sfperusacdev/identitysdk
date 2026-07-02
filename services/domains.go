package services

import (
	"context"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/xreq"
)

func (s *ExternalBridgeService) GetDominios(ctx context.Context) ([]string, error) {
	var apiresponse struct {
		Message string   `json:"message"`
		Data    []string `json:"data"`
	}
	if err := xreq.MakeRequest(ctx,
		identitysdk.GetIdentityServer(),
		"/v1/get-list-empresas",
		xreq.WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return nil, err
	}
	return apiresponse.Data, nil
}

type Empresa struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// Deprecated: use bridge.Identity.GetEmpresasV2 instead.
func (s *ExternalBridgeService) GetEmpresas(ctx context.Context) ([]Empresa, error) {
	var apiresponse struct {
		Message string    `json:"message"`
		Data    []Empresa `json:"data"`
	}
	if err := xreq.MakeRequest(ctx,
		identitysdk.GetIdentityServer(),
		"/v1/get-list-empresas-full",
		xreq.WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return nil, err
	}
	return apiresponse.Data, nil
}
