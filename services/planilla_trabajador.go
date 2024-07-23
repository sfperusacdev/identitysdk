package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/entities"
)

func (s *ExternalBridgeService) GetPlanillaTrabajador(ctx context.Context, trabajadorCodigo string) (*entities.TrabajadorPlanilla, error) {
	var company, token = s.readCompanyAndToken(ctx)
	baseurl, err := identitysdk.GetContratosServiceURL(ctx, company)
	if err != nil {
		slog.Error("error trying to retrieve `contratos` service url", "error", err)
		return nil, err
	}
	var apiresponse struct {
		Message string                      `json:"message"`
		Data    entities.TrabajadorPlanilla `json:"data"`
	}

	var enpointPath = fmt.Sprintf("/v1/trabajadores/planilla/%s/%s", company, identitysdk.RemovePrefix(trabajadorCodigo))
	if err := s.makeRequest(ctx, baseurl, enpointPath, token, &apiresponse); err != nil {
		return nil, err
	}
	return &apiresponse.Data, nil
}
