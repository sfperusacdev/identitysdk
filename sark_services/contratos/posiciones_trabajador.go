package contratos

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/sfperusacdev/identitysdk/xreq"
)

var ErrNotFound = errors.New("the record you are looking for was not found")

func (s *ContratosService) GetTrabajadorPosicion(ctx context.Context, codigo string) (string, error) {
	baseURL, err := s.env.GetContratosServiceURL(ctx, s.env.Empresa(ctx))
	if err != nil {
		slog.Error("error trying to retrieve `contratos` service url", "error", err)
		return "", err
	}
	var response struct {
		Data []struct {
			CargoCodigo string `json:"cargo_codigo"`
		} `json:"data"`
	}
	if err := xreq.MakeRequest(ctx, baseURL, fmt.Sprintf("/v1/fotocheck/%s/json", codigo),
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
		xreq.WithUnmarshalResponseInto(&response),
	); err != nil {
		return "", err
	}
	if len(response.Data) == 0 {
		return "", ErrNotFound
	}
	return response.Data[0].CargoCodigo, nil
}
