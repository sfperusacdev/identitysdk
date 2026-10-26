package asistencia

import (
	"context"
	"time"

	asistenciapb "github.com/sfperusacdev/identitysdk/grpc/gen/asistencia"
	"github.com/sfperusacdev/identitysdk/telemetry"
	"github.com/sfperusacdev/identitysdk/utils/ranges"
	"github.com/shopspring/decimal"
	"github.com/user0608/goones/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type ResultTrabajadoresHorarios struct {
	TrabajadorCodigo string `json:"trabajador_codigo"`
	Fechas           []Dia  `json:"fechas"`
}

type Dia struct {
	Fecha             time.Time       `json:"fecha"`
	HorasRegularesDia decimal.Decimal `json:"horas_regulares_dia"`
	Horarios          []Rango         `json:"horarios"`
	Descansos         []RangoDescanso `json:"descansos"`
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

func (s *AsistenciaService) ResumenHorarios_RangoJornales(ctx context.Context,
	desde, hasta types.DateOnly, trabajadores []string) (_ []ResultTrabajadoresHorarios, err error) {
	tracer := telemetry.Current().TracerProvider().Tracer("identitysdk/sark_services/asistencia")
	ctx, span := tracer.Start(ctx, "asistencia.resumen_horarios")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, "")
		}
		span.End()
	}()
	span.SetAttributes(attribute.Int("horarios.trabajadores_solicitados", len(trabajadores)))
	rpcCtx, rpcSpan := tracer.Start(ctx, "asistencia.resumen_horarios.rpc")
	results, err := s.asistenciaGrpc.ListarTrabajadoresHorarios(rpcCtx, &asistenciapb.ListarTrabajadoresHorariosRequest{
		FechaInicio:         desde.String(),
		FechaFin:            hasta.String(),
		TrabajadoresCodigos: trabajadores,
	}, grpc.MaxCallRecvMsgSize(16*1024*1024))
	if err != nil {
		rpcSpan.SetStatus(codes.Error, "")
	}
	rpcSpan.End()
	if err != nil {
		return nil, err
	}
	_, conversionSpan := tracer.Start(ctx, "asistencia.resumen_horarios.convertir")
	defer conversionSpan.End()
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("horarios.respuesta_bytes", proto.Size(results)),
			attribute.Int("horarios.trabajadores_devueltos", len(results.GetData())))
	}

	response := make([]ResultTrabajadoresHorarios, 0, len(results.GetData()))
	for _, result := range results.GetData() {
		trabajador := ResultTrabajadoresHorarios{
			TrabajadorCodigo: result.GetTrabajadorCodigo(),
			Fechas:           make([]Dia, 0, len(result.GetFechas())),
		}

		for _, fecha := range result.GetFechas() {
			horasRegularesDia, err := decimal.NewFromString(fecha.GetHorasRegularesDia())
			if err != nil {
				conversionSpan.SetStatus(codes.Error, "")
				return nil, err
			}

			dia := Dia{
				HorasRegularesDia: horasRegularesDia,
				Horarios:          make([]Rango, 0, len(fecha.GetHorarios())),
				Descansos:         make([]RangoDescanso, 0, len(fecha.GetDescansos())),
			}
			if timestamp := fecha.GetFecha(); timestamp != nil {
				dia.Fecha = timestamp.AsTime()
			}

			for _, horario := range fecha.GetHorarios() {
				rango := Rango{}
				if timestamp := horario.GetInicio(); timestamp != nil {
					rango.Inicio = timestamp.AsTime()
				}
				if timestamp := horario.GetFin(); timestamp != nil {
					rango.Fin = timestamp.AsTime()
				}
				if timestamp := horario.GetInicioWindows(); timestamp != nil {
					rango.InicioWindows = timestamp.AsTime()
				}
				if timestamp := horario.GetFinWindows(); timestamp != nil {
					rango.FinWindows = timestamp.AsTime()
				}

				dia.Horarios = append(dia.Horarios, rango)
			}

			for _, descanso := range fecha.GetDescansos() {
				rango := RangoDescanso{}
				if timestamp := descanso.GetInicio(); timestamp != nil {
					rango.Inicio = timestamp.AsTime()
				}
				if timestamp := descanso.GetFin(); timestamp != nil {
					rango.Fin = timestamp.AsTime()
				}

				dia.Descansos = append(dia.Descansos, rango)
			}

			trabajador.Fechas = append(trabajador.Fechas, dia)
		}

		response = append(response, trabajador)
	}

	return response, nil
}
