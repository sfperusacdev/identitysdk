package identity

import (
	"context"
	"fmt"
	"time"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/xreq"
)

type IdentityService struct {
	identity IdentityProvider
}

func NewIdentityService(identity IdentityProvider) *IdentityService {
	return &IdentityService{identity: identity}
}

func (s *IdentityService) GetDominios(ctx context.Context) ([]string, error) {
	var apiresponse struct {
		Message string   `json:"message"`
		Data    []string `json:"data"`
	}
	if err := xreq.MakeRequest(ctx,
		s.identity.IdentityServer(),
		"/v1/get-list-empresas",
		xreq.WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return nil, err
	}
	return apiresponse.Data, nil
}

type SucursalDto struct {
	Code         string  `json:"code" chk:"nonil"`
	Description  string  `json:"description" chk:"nonil"`
	Address      string  `json:"address"`
	ExternalReff *string `json:"external_reff"`
	CompanyCode  string  `json:"company_code"`
	IsDisabled   bool    `json:"is_disabled"`

	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
	WriteAt   time.Time `json:"write_at"`
	WriteBy   string    `json:"write_by"`
}

type EmpresaDto struct {
	Code           string        `json:"code" chk:"nonil"`
	Description    string        `json:"description" chk:"nonil"`
	BusinessName   string        `json:"business_name"`
	BusinessDoc    string        `json:"business_doc"`
	Address        string        `json:"address"`
	IsDisabled     bool          `json:"is_disabled"`
	Comment        *string       `json:"comment"`
	ImageLocation  *string       `json:"image_location"`
	ExternalReff   *string       `json:"external_reff"`
	IntegrationUrl *string       `json:"integration_url"`
	Zona           string        `json:"zona"`
	Sucursales     []SucursalDto `json:"sucursales"`
	CreatedAt      time.Time     `json:"created_at"`
	CreatedBy      string        `json:"created_by"`
	WriteAt        time.Time     `json:"write_at"`
	WriteBy        string        `json:"write_by"`
}

// GetEmpresas retrieves the list of companies from the identity service.
// Requires a valid access token for authentication.
func (s *IdentityService) GetEmpresas(ctx context.Context) ([]EmpresaDto, error) {
	var apiresponse struct {
		Message string       `json:"message"`
		Data    []EmpresaDto `json:"data"`
	}
	if err := xreq.MakeRequest(ctx,
		s.identity.IdentityServer(),
		"/v1/internal/get-list-empresas",
		xreq.WithUnmarshalResponseInto(&apiresponse),
		xreq.WithAccessToken(s.identity.AccessToken()),
	); err != nil {
		return nil, err
	}
	for i := range apiresponse.Data {
		for j := range apiresponse.Data[i].Sucursales {
			apiresponse.Data[i].Sucursales[j].Code =
				identitysdk.RemovePrefix(apiresponse.Data[i].Sucursales[j].Code)
		}
	}
	return apiresponse.Data, nil
}

// GetEmpresa retrieves the details of a single company identified by its code from the identity service.
// Requires a valid access token for authentication.
func (s *IdentityService) GetEmpresa(ctx context.Context, domain string) (*EmpresaDto, error) {
	var apiresponse struct {
		Message string      `json:"message"`
		Data    *EmpresaDto `json:"data"`
	}
	if err := xreq.MakeRequest(ctx,
		s.identity.IdentityServer(),
		fmt.Sprintf("/v1/internal/get-list-empresas/%s", domain),
		xreq.WithUnmarshalResponseInto(&apiresponse),
		xreq.WithAccessToken(s.identity.AccessToken()),
	); err != nil {
		return nil, err
	}
	if apiresponse.Data != nil {
		for j := range apiresponse.Data.Sucursales {
			apiresponse.Data.Sucursales[j].Code =
				identitysdk.RemovePrefix(apiresponse.Data.Sucursales[j].Code)
		}
	}
	return apiresponse.Data, nil
}
