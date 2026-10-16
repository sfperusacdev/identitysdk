package contratos

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/entities"
	"github.com/sfperusacdev/identitysdk/xreq"
)

type PlanillaDto struct {
	Codigo                      string `json:"codigo"`
	Descripcion                 string `json:"descripcion"`
	DefaultGrupoProvisionCodigo string `json:"default_grupo_provision_codigo"`
}

func (s *ContratosService) GetPlanillas(ctx context.Context) ([]PlanillaDto, error) {
	baseURL, err := s.env.GetContratosServiceURL(ctx, s.env.Empresa(ctx))
	if err != nil {
		slog.Error("error trying to retrieve `contratos` service url", "error", err)
		return nil, err
	}
	var response struct {
		Data []PlanillaDto `json:"data"`
	}
	if err := xreq.MakeRequest(ctx, baseURL, "/v1/api/tables/empresa/planilla",
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
		xreq.WithUnmarshalResponseInto(&response),
	); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (s *ContratosService) GetPlanillaTrabajador(ctx context.Context, trabajadorCodigo string) (*entities.TrabajadorPlanilla, error) {
	company := s.env.Empresa(ctx)
	baseURL, err := s.env.GetContratosServiceURL(ctx, company)
	if err != nil {
		slog.Error("error trying to retrieve `contratos` service url", "error", err)
		return nil, err
	}
	var response struct {
		Data entities.TrabajadorPlanilla `json:"data"`
	}
	endpoint := fmt.Sprintf("/v1/trabajadores/planilla/%s/%s", company, identitysdk.RemovePrefix(trabajadorCodigo))
	if err := xreq.MakeRequest(ctx, baseURL, endpoint,
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
		xreq.WithUnmarshalResponseInto(&response),
	); err != nil {
		return nil, err
	}
	return &response.Data, nil
}
