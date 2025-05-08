package services

import (
	"context"
	"fmt"
	"time"

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

// Deprecated: use GetEmpresasV2 instead.
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

// v2

type EmpresaDto struct {
	Code           string    `json:"code" chk:"nonil"`
	Description    string    `json:"description" chk:"nonil"`
	BusinessName   string    `json:"business_name"`
	BusinessDoc    string    `json:"business_doc"`
	Address        string    `json:"address"`
	IsDisabled     bool      `json:"is_disabled"`
	Comment        *string   `json:"comment"`
	ImageLocation  *string   `json:"image_location"`
	ExternalReff   *string   `json:"external_reff"`
	IntegrationUrl *string   `json:"integration_url"`
	Zona           string    `json:"zona"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      string    `json:"created_by"`
	WriteAt        time.Time `json:"write_at"`
	WriteBy        string    `json:"write_by"`
}

// GetEmpresasV2 retrieves the list of companies from the identity service.
// Requires a valid access token for authentication.
func (s *ExternalBridgeService) GetEmpresasV2(ctx context.Context) ([]EmpresaDto, error) {
	var apiresponse struct {
		Message string       `json:"message"`
		Data    []EmpresaDto `json:"data"`
	}
	if err := xreq.MakeRequest(ctx,
		identitysdk.GetIdentityServer(),
		"/v1/internal/get-list-empresas",
		xreq.WithUnmarshalResponseInto(&apiresponse),
		xreq.WithAccessToken(identitysdk.GetAccessToken()),
	); err != nil {
		return nil, err
	}
	return apiresponse.Data, nil
}

// GetEmpresaV2 retrieves the details of a single company identified by its code from the identity service.
// Requires a valid access token for authentication.
func (s *ExternalBridgeService) GetEmpresaV2(ctx context.Context, codigo string) (*EmpresaDto, error) {
	var apiresponse struct {
		Message string      `json:"message"`
		Data    *EmpresaDto `json:"data"`
	}
	if err := xreq.MakeRequest(ctx,
		identitysdk.GetIdentityServer(),
		fmt.Sprintf("/v1/internal/get-list-empresas/%s", codigo),
		xreq.WithUnmarshalResponseInto(&apiresponse),
		xreq.WithAccessToken(identitysdk.GetAccessToken()),
	); err != nil {
		return nil, err
	}
	return apiresponse.Data, nil
}
