package services

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/utils/turnos"
	"github.com/sfperusacdev/identitysdk/xreq"
)

func (s *ExternalBridgeService) Turnos(ctx context.Context) ([]turnos.Turno, error) {
	company, token := s.readCompanyAndToken(ctx)
	baseURL, err := identitysdk.GetAsistenciaServiceURL(ctx, company)
	if err != nil {
		slog.Error("error retrieving service URL", "error", err)
		return nil, err
	}

	var response struct {
		Message string `json:"message"`
		Data    []struct {
			Codigo            string `json:"codigo"`
			Descripcion       string `json:"descripcion"`
			Inicio            string `json:"inicio"`
			Fin               string `json:"fin"`
			ReferenciaExterna string `json:"referecia_externa"`
		} `json:"data"`
	}

	endpoint := "v1/api/turnos"

	if err := xreq.MakeRequest(ctx,
		baseURL, endpoint,
		xreq.WithAuthorization(token),
		xreq.WithUnmarshalResponseInto(&response),
		xreq.WithQueryParams(url.Values{"utc": []string{"1"}}),
	); err != nil {
		return nil, err
	}

	result := make([]turnos.Turno, 0, len(response.Data))
	for _, itm := range response.Data {
		result = append(result, turnos.Turno{
			Codigo:           itm.Codigo,
			Descripcion:      itm.Descripcion,
			Inicio:           itm.Inicio,
			Fin:              itm.Fin,
			RefereciaExterna: itm.ReferenciaExterna,
		})
	}

	return turnos.ValidarTurnos(result)
}
