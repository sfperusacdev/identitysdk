package asistencia

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/sfperusacdev/identitysdk/utils/turnos"
	"github.com/sfperusacdev/identitysdk/xreq"
	"github.com/user0608/goones/types"
)

func (s *AsistenciaService) Turnos(ctx context.Context) ([]turnos.Turno, error) {
	baseURL, err := s.env.GetAsistenciaServiceURL(ctx, s.env.Empresa(ctx))
	if err != nil {
		slog.Error("error retrieving service URL", "error", err)
		return nil, err
	}
	var response struct {
		Data []struct {
			Codigo            string         `json:"codigo"`
			Descripcion       string         `json:"descripcion"`
			Inicio            types.JustTime `json:"inicio"`
			Fin               types.JustTime `json:"fin"`
			ReferenciaExterna string         `json:"referecia_externa"`
		} `json:"data"`
	}
	if err := xreq.MakeRequest(ctx, baseURL, "v1/api/turnos",
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
		xreq.WithUnmarshalResponseInto(&response),
		xreq.WithQueryParams(url.Values{"utc": []string{"1"}}),
	); err != nil {
		return nil, err
	}
	result := make([]turnos.Turno, 0, len(response.Data))
	for _, item := range response.Data {
		result = append(result, turnos.Turno{
			Codigo:           item.Codigo,
			Descripcion:      item.Descripcion,
			Inicio:           item.Inicio,
			Fin:              item.Fin,
			RefereciaExterna: item.ReferenciaExterna,
		})
	}
	return turnos.ValidarTurnos(result)
}
