package global

import (
	"context"
	"fmt"
	"log/slog"

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

func (s *GlobalService) GetUbigeos(ctx context.Context) ([]Ubigeo, error) {
	baseURL, err := s.env.GetGeneralServiceURL(ctx, s.env.Empresa(ctx))
	if err != nil {
		slog.Error("error retrieving service URL", "error", err)
		return nil, err
	}
	var response struct {
		Data []Ubigeo `json:"data"`
	}
	if err := xreq.MakeRequest(ctx, baseURL, "/api/v2/ubigeos",
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
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

func (s *GlobalService) SearchUbigeo(ctx context.Context, source UbigeoSource, departmentName, provinceName, districtName string) (*Ubigeo, error) {
	baseURL, err := s.env.GetGeneralServiceURL(ctx, s.env.Empresa(ctx))
	if err != nil {
		slog.Error("error retrieving general service URL", "error", err)
		return nil, err
	}
	var response struct {
		Data *Ubigeo `json:"data"`
	}
	endpoint := fmt.Sprintf("/api/v2/ubigeos/%s/%s/%s/%s", source, departmentName, provinceName, districtName)
	if err := xreq.MakeRequest(ctx, baseURL, endpoint,
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
		xreq.WithUnmarshalResponseInto(&response),
		xreq.WithJsonContentType(),
	); err != nil {
		return nil, err
	}
	return response.Data, nil
}
