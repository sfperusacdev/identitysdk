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

func (s *ExternalBridgeService) GetCadenaOrganigrama(ctx context.Context, posicion string) ([]organigramaResult, error) {
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
	var enpointPath = path.Join("/v1/api/organigrama/cadena_organigrama/", posicion)
	if err := s.makeRequest(ctx, baseurl, enpointPath, token, &apiresponse); err != nil {
		return nil, err
	}
	return apiresponse.Data, nil
}
