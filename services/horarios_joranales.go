package services

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/utils/ranges"
	"github.com/sfperusacdev/identitysdk/xreq"
	"github.com/user0608/goones/types"
)

type ResultTrabajadoresHorarios struct {
	TrabajadorCodigo string `json:"trabajador_codigo"`
	Fechas           []Dia  `json:"fechas"`
}

type Dia struct {
	Fecha     time.Time       `json:"fecha"`
	Horarios  []Rango         `json:"horarios"`
	Descansos []RangoDescanso `json:"descansos"`
}

type Rango struct {
	Fecha         time.Time `json:"-"`
	Inicio        time.Time `json:"inicio"`
	Fin           time.Time `json:"fin"`
	InicioWindows time.Time `json:"inicio_windows"`
	FinWindows    time.Time `json:"fin_windows"`
}

var _ ranges.TimeRange = (*Rango)(nil)
var _ ranges.TimeWindowsRange = (*Rango)(nil)

func (r Rango) StartTime() time.Time { return r.Inicio }
func (r Rango) EndTime() time.Time   { return r.Fin }

func (r Rango) StartWindowsTime() time.Time { return r.InicioWindows }
func (r Rango) EndWindowsTime() time.Time   { return r.FinWindows }

type RangoDescanso struct {
	Inicio time.Time `json:"inicio"`
	Fin    time.Time `json:"fin"`
}

var _ ranges.TimeRange = (*RangoDescanso)(nil)

func (r RangoDescanso) StartTime() time.Time { return r.Inicio }

func (r RangoDescanso) EndTime() time.Time { return r.Fin }

func (s *ExternalBridgeService) ResumenHorarios_RangoJornales(ctx context.Context,
	desde, hasta types.DateOnly, trabajadores []string) ([]ResultTrabajadoresHorarios, error) {
	company, token := s.readCompanyAndToken(ctx)
	baseURL, err := identitysdk.GetAsistenciaServiceURL(ctx, company)
	if err != nil {
		slog.Error("error retrieving service URL", "error", err)
		return nil, err
	}

	var response struct {
		Message string                       `json:"message"`
		Data    []ResultTrabajadoresHorarios `json:"data"`
	}

	if err := xreq.MakeRequest(ctx,
		baseURL, "v1/api/horarios/resumen",
		xreq.WithMethod(http.MethodPost),
		xreq.WithAuthorization(token),
		xreq.WithQueryParam("desde", desde.String()),
		xreq.WithQueryParam("hasta", hasta.String()),
		xreq.WithJSONBody(map[string]any{"values": trabajadores}),
		xreq.WithUnmarshalResponseInto(&response),
	); err != nil {
		return nil, err
	}
	return response.Data, nil
}
