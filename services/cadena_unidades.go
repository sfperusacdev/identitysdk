package services

import (
	"context"
	"log/slog"
	"path"

	"github.com/sfperusacdev/identitysdk"
)

type unidadesResult struct {
	Codigo      string `json:"codigo"`
	Descripcion string `json:"descripcion"`
}

func (s *ExternalBridgeService) GetCadenaDeUnidades(ctx context.Context, posicion string) ([]unidadesResult, error) {
	var company, token = s.readCompanyAndToken(ctx)
	baseurl, err := identitysdk.GetGeneralServiceURL(ctx, company)
	if err != nil {
		slog.Error("error trying to retrieve `contratos` service url", "error", err)
		return nil, err
	}
	var apiresponse struct {
		Message string           `json:"message"`
		Data    []unidadesResult `json:"data"`
	}
	var enpointPath = path.Join("/v1/api/organigrama/cadena_unidades/", posicion)
	if err := s.makeRequest(ctx, baseurl, enpointPath, token, &apiresponse); err != nil {
		return nil, err
	}
	return apiresponse.Data, nil
}
