package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/xreq"
)

type Ubigeo struct {
	Codigo                   string  `gorm:"primaryKey" json:"codigo"`
	DepartamentoIneiCodigo   *string `json:"departamento_inei_codigo"`
	DepartamentoIneiNombre   *string `json:"departamento_inei_nombre"`
	ProvinciaIneiCodigo      *string `json:"provincia_inei_codigo"`
	ProvinciaIneiNombre      *string `json:"provincia_inei_nombre"`
	DistritoIneiCodigo       *string `json:"distrito_inei_codigo"`
	DistritoIneiNombre       *string `json:"distrito_inei_nombre"`
	DepartamentoReniecCodigo *string `json:"departamento_reniec_codigo"`
	DepartamentoReniecNombre *string `json:"departamento_reniec_nombre"`
	ProvinciaReniecCodigo    *string `json:"provincia_reniec_codigo"`
	ProvinciaReniecNombre    *string `json:"provincia_reniec_nombre"`
	DistritoReniecCodigo     *string `json:"distrito_reniec_codigo"`
	DistritoReniecNombre     *string `json:"distrito_reniec_nombre"`
	DepartamentoSunatCodigo  *string `json:"departamento_sunat_codigo"`
	DepartamentoSunatNombre  *string `json:"departamento_sunat_nombre"`
	ProvinciaSunatCodigo     *string `json:"provincia_sunat_codigo"`
	ProvinciaSunatNombre     *string `json:"provincia_sunat_nombre"`
	DistritoSunatCodigo      *string `json:"distrito_sunat_codigo"`
	DistritoSunatNombre      *string `json:"distrito_sunat_nombre"`
}

func (s *ExternalBridgeService) GetUbigeos(ctx context.Context) ([]Ubigeo, error) {
	company, token := s.readCompanyAndToken(ctx)
	baseURL, err := identitysdk.GetGeneralServiceURL(ctx, company)
	if err != nil {
		slog.Error("error retrieving service URL", "error", err)
		return nil, err
	}

	var response struct {
		Message string   `json:"message"`
		Data    []Ubigeo `json:"data"`
	}

	if err := xreq.MakeRequest(ctx,
		baseURL, "/api/v2/ubigeos",
		xreq.WithAuthorization(token),
		xreq.WithUnmarshalResponseInto(&response),
		xreq.WithJsonContentType(),
	); err != nil {
		return nil, err
	}

	return response.Data, nil
}

type UbigeoSource string

const (
	UbigeoSourceINEI   UbigeoSource = "inei"
	UbigeoSourceRENIEC UbigeoSource = "reniec"
	UbigeoSourceSUNAT  UbigeoSource = "sunat"
)

func (s *ExternalBridgeService) SearchUbigeo(
	ctx context.Context,
	source UbigeoSource,
	departmentName string,
	provinceName string,
	districtName string,
) (*Ubigeo, error) {
	company, token := s.readCompanyAndToken(ctx)

	serviceURL, err := identitysdk.GetGeneralServiceURL(ctx, company)
	if err != nil {
		slog.Error("error retrieving general service URL", "error", err)
		return nil, err
	}

	var result struct {
		Message string  `json:"message"`
		Data    *Ubigeo `json:"data"`
	}

	route := fmt.Sprintf(
		"/api/v2/ubigeos/%s/%s/%s/%s",
		string(source),
		departmentName,
		provinceName,
		districtName,
	)

	if err := xreq.MakeRequest(
		ctx,
		serviceURL,
		route,
		xreq.WithAuthorization(token),
		xreq.WithUnmarshalResponseInto(&result),
		xreq.WithJsonContentType(),
	); err != nil {
		return nil, err
	}

	return result.Data, nil
}
