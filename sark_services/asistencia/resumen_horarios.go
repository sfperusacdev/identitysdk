package asistencia

import (
	"context"
	"time"

	asistenciapb "github.com/sfperusacdev/identitysdk/grpc/gen/asistencia"
	"github.com/sfperusacdev/identitysdk/utils/ranges"
	"github.com/shopspring/decimal"
	"github.com/user0608/goones/types"
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
	desde, hasta types.DateOnly, trabajadores []string) ([]ResultTrabajadoresHorarios, error) {
	results, err := s.asistenciaGrpc.ListarTrabajadoresHorarios(ctx, &asistenciapb.ListarTrabajadoresHorariosRequest{
		FechaInicio:         desde.String(),
		FechaFin:            hasta.String(),
		TrabajadoresCodigos: trabajadores,
	})
	if err != nil {
		return nil, err
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
