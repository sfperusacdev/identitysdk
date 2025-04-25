package services

import (
	"context"
	"log/slog"
	"path"

	"github.com/sfperusacdev/identitysdk"
)

type organigramaResult struct {
	Codigo                  string  `json:"codigo"`
	OrganigramaCodigo       string  `json:"organigrama_codigo"`
	Descripcion             string  `json:"descripcion"`
	OrganigramaObjectCodigo *string `json:"organigrama_object_codigo"`
	TypeObject              string  `json:"type_object"`
	Codificacion            *string `json:"codificacion"`
	FuncionCodigo           *string `json:"funcion_codigo"`
}

func (s *ExternalBridgeService) GetPuestosUnidadesSuperiores(ctx context.Context, posicion string) ([]organigramaResult, error) {
	var company, token = s.readCompanyAndToken(ctx)
	baseurl, err := identitysdk.GetGeneralServiceURL(ctx, company)
	if err != nil {
		slog.Error("error trying to retrieve `contratos` service url", "error", err)
		return nil, err
	}
	var apiresponse struct {
		Message string              `json:"message"`
		Data    []organigramaResult `json:"data"`
	}
	var enpointPath = path.Join("/v1/api/organigrama/cadena_organigrama_superiores/", posicion)

	if err := s.MakeRequest(ctx,
		baseurl, enpointPath,
		WithAuthorization(token),
		WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return nil, err
	}
	return apiresponse.Data, nil
}

func (s *ExternalBridgeService) GetPuestosSuperiores(ctx context.Context, posicion string) ([]string, error) {
	records, err := s.GetPuestosUnidadesSuperiores(ctx, posicion)
	if err != nil {
		return nil, err
	}
	var items = []string{}
	for _, r := range records {
		if r.TypeObject == "posicion" && r.Codificacion != nil {
			items = append(items, *r.Codificacion)
		}
	}
	return items, nil

}

type unidadesResult struct {
	Codigo      string `json:"codigo"`
	Descripcion string `json:"descripcion"`
}

func (s *ExternalBridgeService) GetUnidadesSuperiores(ctx context.Context, posicion string) ([]unidadesResult, error) {
	records, err := s.GetPuestosUnidadesSuperiores(ctx, posicion)
	if err != nil {
		return nil, err
	}
	var items = []unidadesResult{}
	for _, r := range records {
		if r.TypeObject == "unidad-organizativa" {
			items = append(items, unidadesResult{Codigo: r.Codigo, Descripcion: r.Descripcion})
		}
	}
	return items, nil
}

func (s *ExternalBridgeService) GetPuestosUnidadesInferiores(ctx context.Context, posicion string) ([]organigramaResult, error) {
	var company, token = s.readCompanyAndToken(ctx)
	baseurl, err := identitysdk.GetGeneralServiceURL(ctx, company)
	if err != nil {
		slog.Error("error trying to retrieve `contratos` service url", "error", err)
		return nil, err
	}
	var apiresponse struct {
		Message string              `json:"message"`
		Data    []organigramaResult `json:"data"`
	}
	var enpointPath = path.Join("/v1/api/organigrama/cadena_organigrama_subordinados/", posicion)

	if err := s.MakeRequest(ctx,
		baseurl, enpointPath,
		WithAuthorization(token),
		WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return nil, err
	}
	return apiresponse.Data, nil
}

func (s *ExternalBridgeService) GetPuestosInferiores(ctx context.Context, posicion string) ([]string, error) {
	records, err := s.GetPuestosUnidadesInferiores(ctx, posicion)
	if err != nil {
		return nil, err
	}
	var items = []string{}
	for _, r := range records {
		if r.TypeObject == "posicion" && r.Codificacion != nil {
			items = append(items, *r.Codificacion)
		}
	}
	return items, nil

}

func (s *ExternalBridgeService) GetUnidadesInferiores(ctx context.Context, posicion string) ([]unidadesResult, error) {
	records, err := s.GetPuestosUnidadesInferiores(ctx, posicion)
	if err != nil {
		return nil, err
	}
	var items = []unidadesResult{}
	for _, r := range records {
		if r.TypeObject == "unidad-organizativa" {
			items = append(items, unidadesResult{Codigo: r.Codigo, Descripcion: r.Descripcion})
		}
	}
	return items, nil
}
