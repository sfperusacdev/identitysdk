package contratos

import (
	"context"
	"log/slog"

	"github.com/sfperusacdev/identitysdk/xreq"
)

func (s *ContratosService) GetPosicionesActivas(ctx context.Context) ([]string, error) {
	baseURL, err := s.env.GetContratosServiceURL(ctx, s.env.Empresa(ctx))
	if err != nil {
		slog.Error("error trying to retrieve `contratos` service url", "error", err)
		return nil, err
	}
	var response struct {
		Data []string `json:"data"`
	}
	if err := xreq.MakeRequest(ctx, baseURL, "/v1/fotocheck/resumen/puestos",
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
		xreq.WithUnmarshalResponseInto(&response),
	); err != nil {
		return nil, err
	}
	return response.Data, nil
}
