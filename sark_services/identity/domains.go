package identity

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/configs"
	"github.com/sfperusacdev/identitysdk/xreq"
	"github.com/user0608/goones/errs"
)

type IdentityService struct {
	conf     configs.GeneralServiceConfigProvider
	identity IdentityProvider
}

func NewIdentityService(conf configs.GeneralServiceConfigProvider, identity IdentityProvider) *IdentityService {
	return &IdentityService{conf: conf, identity: identity}
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

func (s *IdentityService) GetServiceDominios(ctx context.Context) ([]string, error) {
	var apiresponse struct {
		Message string   `json:"message"`
		Data    []string `json:"data"`
	}

	if err := xreq.MakeRequest(ctx,
		s.identity.IdentityServer(),
		"/v1/get-list-empresas",
		xreq.WithUnmarshalResponseInto(&apiresponse),
		xreq.WithQueryParam("resource_code", s.conf.ServiceID()),
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

func (e *EmpresaDto) Tz() (*time.Location, error) {
	e.Zona = strings.TrimSpace(e.Zona)
	if e.Zona == "" {
		return nil, errs.InternalErrorDirect("zona horaria no definida para este dominio")
	}
	location, err := time.LoadLocation(e.Zona)
	if err != nil {
		return nil, err
	}
	return location, nil
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

	if apiresponse.Data == nil {
		return nil, errs.NotFoundDirect("dominio no encontado")
	}

	for j := range apiresponse.Data.Sucursales {
		apiresponse.Data.Sucursales[j].Code =
			identitysdk.RemovePrefix(apiresponse.Data.Sucursales[j].Code)
	}

	return apiresponse.Data, nil
}

func (s *IdentityService) Tz(ctx context.Context, domain string) (*time.Location, error) {
	empresa, err := s.GetEmpresa(ctx, domain)
	if err != nil {
		return nil, err
	}
	return empresa.Tz()
}
