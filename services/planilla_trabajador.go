package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/entities"
)

type PlanillaDto struct {
	Codigo                      string `json:"codigo"`
	Descripcion                 string `json:"descripcion"`
	DefaultGrupoProvisionCodigo string `json:"default_grupo_provision_codigo"`
}

func (s *ExternalBridgeService) GetPlanillas(ctx context.Context) ([]PlanillaDto, error) {
	var company, token = s.readCompanyAndToken(ctx)
	baseurl, err := identitysdk.GetContratosServiceURL(ctx, company)
	if err != nil {
		slog.Error("error trying to retrieve `contratos` service url", "error", err)
		return nil, err
	}
	var apiresponse struct {
		Message string        `json:"message"`
		Data    []PlanillaDto `json:"data"`
	}

	if err := s.MakeRequest(ctx,
		baseurl, "/v1/api/tables/empresa/planilla",
		WithAuthorization(token),
		WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return nil, err
	}
	return apiresponse.Data, nil
}

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

	if err := s.MakeRequest(ctx,
		baseurl, enpointPath,
		WithAuthorization(token),
		WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return nil, err
	}
	return &apiresponse.Data, nil
}
