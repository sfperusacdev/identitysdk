package services

import (
	"context"
	"log/slog"

	"github.com/sfperusacdev/identitysdk"
)

func (s *ExternalBridgeService) GetPosicionesActivas(ctx context.Context) ([]string, error) {
	var company, token = s.readCompanyAndToken(ctx)
	baseurl, err := identitysdk.GetContratosServiceURL(ctx, company)
	if err != nil {
		slog.Error("error trying to retrieve `contratos` service url", "error", err)
		return nil, err
	}
	var apiresponse struct {
		Message string   `json:"message"`
		Data    []string `json:"data"`
	}
	const enpointPath = "/v1/fotocheck/resumen/puestos"
	if err := s.makeRequest(ctx, baseurl, enpointPath, token, &apiresponse); err != nil {
		return nil, err
	}
	return apiresponse.Data, nil
}
