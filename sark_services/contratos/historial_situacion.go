package contratos

import (
	"context"
	"net/http"
	"time"

	"github.com/sfperusacdev/identitysdk/xreq"
	"github.com/user0608/goones/types"
)

type HistorialSituacionDTO struct {
	Codigo           string     `json:"codigo"`
	Planilla         string     `json:"planilla"`
	TrabajadorCodigo string     `json:"trabajador_codigo"`
	Situacion        string     `json:"situacion"`
	Origen           *string    `json:"origen"`
	OrigenCodigo     *string    `json:"origen_codigo"`
	FechaInicio      time.Time  `json:"fecha_inicio"`
	FechaFin         *time.Time `json:"fecha_fin"`
}

// HistorialSituaciones obtiene el historial de situaciones de trabajadores.
// Cuando activos es true, solicita solo las situaciones vigentes a la fecha
func (s *ContratosService) HistorialSituaciones(ctx context.Context, activos bool) ([]HistorialSituacionDTO, error) {
	apiurl, err := s.env.GetContratosServiceURL(ctx, s.env.Empresa(ctx))
	if err != nil {
		return nil, err
	}

	var apiResponse struct {
		Data []HistorialSituacionDTO `json:"data"`
	}
	situacionActiva := "no"
	if activos {
		situacionActiva = "yes"
	}
	if err := xreq.MakeRequest(ctx, apiurl,
		"/api/v2/trabajadores/historial-situacion",
		xreq.WithQueryParam("sucursal", s.env.Sucursal(ctx)),
		xreq.WithQueryParam("activos", situacionActiva),
		xreq.WithUnmarshalResponseInto(&apiResponse),
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
	); err != nil {
		return nil, err
	}
	return apiResponse.Data, nil
}

// Origen debe identificar la fuente (por ejemplo, asistencia.vacaciones).
// OrigenCodigo admite el código con o sin el prefijo de empresa.
type SincronizarHistorialSituacionInput struct {
	Codigo           string          `json:"codigo"`
	Origen           string          `json:"origen"`
	OrigenCodigo     string          `json:"origen_codigo"`
	TrabajadorCodigo string          `json:"trabajador_codigo"`
	Situacion        string          `json:"situacion"`
	FechaInicio      types.DateOnly  `json:"fecha_inicio"`
	FechaFin         *types.DateOnly `json:"fecha_fin"`
}

func (s *ContratosService) SincronizarHistorialSituacion(ctx context.Context, input SincronizarHistorialSituacionInput) error {
	apiurl, err := s.env.GetContratosServiceURL(ctx, s.env.Empresa(ctx))
	if err != nil {
		return err
	}
	return xreq.MakeRequest(ctx, apiurl, "/api/v2/trabajadores/historial-situacion-origen",
		xreq.WithMethod(http.MethodPut),
		xreq.WithQueryParam("sucursal", s.env.Sucursal(ctx)),
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
		xreq.WithJSONBody(input),
	)
}

// Si el registro ya no existe, la operación también es exitosa.
func (s *ContratosService) EliminarHistorialSituacionPorOrigen(ctx context.Context, origen, origenCodigo string) error {
	apiurl, err := s.env.GetContratosServiceURL(ctx, s.env.Empresa(ctx))
	if err != nil {
		return err
	}
	return xreq.MakeRequest(ctx, apiurl, "/api/v2/trabajadores/historial-situacion-origen",
		xreq.WithMethod(http.MethodDelete),
		xreq.WithQueryParam("sucursal", s.env.Sucursal(ctx)),
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
		xreq.WithJSONBody(struct {
			Origen       string `json:"origen"`
			OrigenCodigo string `json:"origen_codigo"`
		}{Origen: origen, OrigenCodigo: origenCodigo}),
	)
}

func (s *ContratosService) HistorialSituacionesActivas(ctx context.Context, desde, hasta types.DateOnly) ([]HistorialSituacionDTO, error) {
	apiurl, err := s.env.GetContratosServiceURL(ctx, s.env.Empresa(ctx))
	if err != nil {
		return nil, err
	}

	var apiResponse struct {
		Data []HistorialSituacionDTO `json:"data"`
	}

	if err := xreq.MakeRequest(ctx, apiurl,
		"/api/v2/trabajadores/historial-situacion-activos",
		xreq.WithQueryParam("sucursal", s.env.Sucursal(ctx)),
		xreq.WithQueryParam("desde", desde.String()),
		xreq.WithQueryParam("hasta", hasta.String()),
		xreq.WithUnmarshalResponseInto(&apiResponse),
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
	); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}
