package turnos_test

import (
	"testing"
	"time"

	"github.com/sfperusacdev/identitysdk/utils/turnos"
	"github.com/stretchr/testify/require"
)

type TestSegment struct {
	CodigoTurno string
	FechaInicio time.Time
	FechaFin    time.Time
}

func (s TestSegment) GetStartTime() time.Time { return s.FechaInicio }
func (s TestSegment) GetEndTime() time.Time   { return s.FechaFin }
func (s TestSegment) NewSegment(t turnos.Turno, inicio, fin time.Time) turnos.Segmentable {
	return TestSegment{CodigoTurno: t.Codigo, FechaInicio: inicio, FechaFin: fin}
}

func makeDate(valor string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04", valor, time.Local)
	if err != nil {
		panic(err)
	}
	return t
}

func TestTurnoSimple(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	inicio := makeDate("2025-01-01 07:00")
	fin := makeDate("2025-01-01 09:00")

	input := TestSegment{FechaInicio: inicio, FechaFin: fin}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 0)

	require.Len(t, res, 1)

	seg := res[0]

	require.Equal(t, "DAY", seg.CodigoTurno)
	require.Equal(t, inicio, seg.FechaInicio)
	require.Equal(t, fin, seg.FechaFin)
}

func TestCruceDeDosTurnos(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	inicio := makeDate("2025-01-01 17:00")
	fin := makeDate("2025-01-01 19:00")

	input := TestSegment{FechaInicio: inicio, FechaFin: fin}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 0)

	require.Len(t, res, 2)

	esperados := []TestSegment{
		{"DAY", makeDate("2025-01-01 17:00"), makeDate("2025-01-01 18:00")},
		{"NIGHT", makeDate("2025-01-01 18:00"), makeDate("2025-01-01 19:00")},
	}

	for i, esperado := range esperados {
		obt := res[i]

		require.Equal(t, esperado.CodigoTurno, obt.CodigoTurno)
		require.Equal(t, esperado.FechaInicio, obt.FechaInicio)
		require.Equal(t, esperado.FechaFin, obt.FechaFin)
	}
}

func TestTresSegmentos(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "12:00:00"),
		newTurno("AFTERNOON", "12:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 11:00"),
		FechaFin:    makeDate("2025-01-01 19:00"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 0)

	require.Len(t, res, 3)

	esperados := []TestSegment{
		{"DAY", makeDate("2025-01-01 11:00"), makeDate("2025-01-01 12:00")},
		{"AFTERNOON", makeDate("2025-01-01 12:00"), makeDate("2025-01-01 18:00")},
		{"NIGHT", makeDate("2025-01-01 18:00"), makeDate("2025-01-01 19:00")},
	}

	for i := range esperados {
		require.Equal(t, esperados[i].CodigoTurno, res[i].CodigoTurno)
		require.Equal(t, esperados[i].FechaInicio, res[i].FechaInicio)
		require.Equal(t, esperados[i].FechaFin, res[i].FechaFin)
	}
}

func TestResultadoTresSegmentos(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 17:00"),
		FechaFin:    makeDate("2025-01-02 07:00"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 0)

	require.Len(t, res, 3)

	esperados := []TestSegment{
		{"DAY", makeDate("2025-01-01 17:00"), makeDate("2025-01-01 18:00")},
		{"NIGHT", makeDate("2025-01-01 18:00"), makeDate("2025-01-02 06:00")},
		{"DAY", makeDate("2025-01-02 06:00"), makeDate("2025-01-02 07:00")},
	}

	for i := range esperados {
		require.Equal(t, esperados[i].CodigoTurno, res[i].CodigoTurno)
		require.Equal(t, esperados[i].FechaInicio, res[i].FechaInicio)
		require.Equal(t, esperados[i].FechaFin, res[i].FechaFin)
	}
}

func TestRangoDeVariosDias(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 10:00"),
		FechaFin:    makeDate("2025-01-03 09:00"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 0)

	esperados := []TestSegment{
		{"DAY", makeDate("2025-01-01 10:00"), makeDate("2025-01-01 18:00")},
		{"NIGHT", makeDate("2025-01-01 18:00"), makeDate("2025-01-02 06:00")},
		{"DAY", makeDate("2025-01-02 06:00"), makeDate("2025-01-02 18:00")},
		{"NIGHT", makeDate("2025-01-02 18:00"), makeDate("2025-01-03 06:00")},
		{"DAY", makeDate("2025-01-03 06:00"), makeDate("2025-01-03 09:00")},
	}

	require.Len(t, res, len(esperados))

	for i := range esperados {
		require.Equal(t, esperados[i].CodigoTurno, res[i].CodigoTurno)
		require.Equal(t, esperados[i].FechaInicio, res[i].FechaInicio)
		require.Equal(t, esperados[i].FechaFin, res[i].FechaFin)
	}
}

func TestAplicarRangoInput(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 10:00"),
		FechaFin:    makeDate("2025-01-03 09:00"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 0)

	esperados := []TestSegment{
		{"DAY", makeDate("2025-01-01 10:00"), makeDate("2025-01-01 18:00")},
		{"NIGHT", makeDate("2025-01-01 18:00"), makeDate("2025-01-02 06:00")},
		{"DAY", makeDate("2025-01-02 06:00"), makeDate("2025-01-02 18:00")},
		{"NIGHT", makeDate("2025-01-02 18:00"), makeDate("2025-01-03 06:00")},
		{"DAY", makeDate("2025-01-03 06:00"), makeDate("2025-01-03 09:00")},
	}

	require.Len(t, res, len(esperados))

	for i := range esperados {
		require.Equal(t, esperados[i].CodigoTurno, res[i].CodigoTurno)
		require.Equal(t, esperados[i].FechaInicio, res[i].FechaInicio)
		require.Equal(t, esperados[i].FechaFin, res[i].FechaFin)
	}
}

func TestHolguraMin(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 17:00"),
		FechaFin:    makeDate("2025-01-01 18:03"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 5)

	esperados := []TestSegment{
		{"DAY", makeDate("2025-01-01 17:00"), makeDate("2025-01-01 18:03")},
	}

	require.Len(t, res, len(esperados))

	for i := range esperados {
		require.Equal(t, esperados[i].CodigoTurno, res[i].CodigoTurno)
		require.Equal(t, esperados[i].FechaInicio, res[i].FechaInicio)
		require.Equal(t, esperados[i].FechaFin, res[i].FechaFin)
	}
}

func TestInicioExactoEnCambioTurno(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 18:00"),
		FechaFin:    makeDate("2025-01-01 19:00"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 0)

	esperados := []TestSegment{
		{"NIGHT", makeDate("2025-01-01 18:00"), makeDate("2025-01-01 19:00")},
	}

	require.Len(t, res, len(esperados))
	for i := range esperados {
		require.Equal(t, esperados[i].CodigoTurno, res[i].CodigoTurno)
		require.Equal(t, esperados[i].FechaInicio, res[i].FechaInicio)
		require.Equal(t, esperados[i].FechaFin, res[i].FechaFin)
	}
}

func TestFinExactoEnCambioTurno(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 16:00"),
		FechaFin:    makeDate("2025-01-01 18:00"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 0)

	esperados := []TestSegment{
		{"DAY", makeDate("2025-01-01 16:00"), makeDate("2025-01-01 18:00")},
	}

	require.Len(t, res, len(esperados))
	for i := range esperados {
		require.Equal(t, esperados[i].CodigoTurno, res[i].CodigoTurno)
		require.Equal(t, esperados[i].FechaInicio, res[i].FechaInicio)
		require.Equal(t, esperados[i].FechaFin, res[i].FechaFin)
	}
}

func TestTodoDentroDeTurnoCruzado(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("NOCHE", "18:00:00", "06:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 22:00"),
		FechaFin:    makeDate("2025-01-02 03:00"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 0)

	esperados := []TestSegment{
		{"NOCHE", makeDate("2025-01-01 22:00"), makeDate("2025-01-02 03:00")},
	}

	require.Len(t, res, len(esperados))
	for i := range esperados {
		require.Equal(t, esperados[i].CodigoTurno, res[i].CodigoTurno)
		require.Equal(t, esperados[i].FechaInicio, res[i].FechaInicio)
		require.Equal(t, esperados[i].FechaFin, res[i].FechaFin)
	}
}

func TestTurnoCompletoAfueraRango(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 00:10"),
		FechaFin:    makeDate("2025-01-01 05:50"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 0)

	esperados := []TestSegment{
		{"NIGHT", makeDate("2025-01-01 00:10"), makeDate("2025-01-01 05:50")},
	}

	require.Len(t, res, len(esperados))
	for i := range esperados {
		require.Equal(t, esperados[i].CodigoTurno, res[i].CodigoTurno)
		require.Equal(t, esperados[i].FechaInicio, res[i].FechaInicio)
		require.Equal(t, esperados[i].FechaFin, res[i].FechaFin)
	}
}

func TestRangoMinimoDentroDeUnMinuto(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 10:33"),
		FechaFin:    makeDate("2025-01-01 10:34"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 0)

	esperados := []TestSegment{
		{"DAY", makeDate("2025-01-01 10:33"), makeDate("2025-01-01 10:34")},
	}

	require.Len(t, res, 1)
	require.Equal(t, esperados[0].CodigoTurno, res[0].CodigoTurno)
	require.Equal(t, esperados[0].FechaInicio, res[0].FechaInicio)
	require.Equal(t, esperados[0].FechaFin, res[0].FechaFin)
}

func TestHolguraNoAbsorbe(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 17:00"),
		FechaFin:    makeDate("2025-01-01 18:10"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 5)

	esperados := []TestSegment{
		{"DAY", makeDate("2025-01-01 17:00"), makeDate("2025-01-01 18:00")},
		{"NIGHT", makeDate("2025-01-01 18:00"), makeDate("2025-01-01 18:10")},
	}

	require.Len(t, res, len(esperados))
	for i := range esperados {
		require.Equal(t, esperados[i].CodigoTurno, res[i].CodigoTurno)
		require.Equal(t, esperados[i].FechaInicio, res[i].FechaInicio)
		require.Equal(t, esperados[i].FechaFin, res[i].FechaFin)
	}
}

func TestHolguraExactaAbsorbe(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 17:00"),
		FechaFin:    makeDate("2025-01-01 18:05"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 5)

	esperados := []TestSegment{
		{"DAY", makeDate("2025-01-01 17:00"), makeDate("2025-01-01 18:05")},
	}

	require.Len(t, res, 1)

	require.Equal(t, esperados[0].CodigoTurno, res[0].CodigoTurno)
	require.Equal(t, esperados[0].FechaInicio, res[0].FechaInicio)
	require.Equal(t, esperados[0].FechaFin, res[0].FechaFin)
}

func TestHolguraJustoInsuficiente(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 17:00"),
		FechaFin:    makeDate("2025-01-01 18:06"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 5)

	esperados := []TestSegment{
		{"DAY", makeDate("2025-01-01 17:00"), makeDate("2025-01-01 18:00")},
		{"NIGHT", makeDate("2025-01-01 18:00"), makeDate("2025-01-01 18:06")},
	}

	require.Len(t, res, 2)

	for i := range esperados {
		require.Equal(t, esperados[i].CodigoTurno, res[i].CodigoTurno)
		require.Equal(t, esperados[i].FechaInicio, res[i].FechaInicio)
		require.Equal(t, esperados[i].FechaFin, res[i].FechaFin)
	}
}

func TestHolguraMuyGrandeAbsorbeTodoSiguienteTurno(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 17:00"),
		FechaFin:    makeDate("2025-01-01 18:10"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 60)

	esperados := []TestSegment{
		{"DAY", makeDate("2025-01-01 17:00"), makeDate("2025-01-01 18:10")},
	}

	require.Len(t, res, 1)
	require.Equal(t, esperados[0].CodigoTurno, res[0].CodigoTurno)
	require.Equal(t, esperados[0].FechaInicio, res[0].FechaInicio)
	require.Equal(t, esperados[0].FechaFin, res[0].FechaFin)
}

func TestHolguraSegmentoCruzaMedianoche(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 17:00"),
		FechaFin:    makeDate("2025-01-02 00:05"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 10)

	esperados := []TestSegment{
		{"DAY", makeDate("2025-01-01 17:00"), makeDate("2025-01-01 18:00")},
		{"NIGHT", makeDate("2025-01-01 18:00"), makeDate("2025-01-02 00:05")},
	}

	require.Len(t, res, 2)
	for i := range esperados {
		require.Equal(t, esperados[i].CodigoTurno, res[i].CodigoTurno)
		require.Equal(t, esperados[i].FechaInicio, res[i].FechaInicio)
		require.Equal(t, esperados[i].FechaFin, res[i].FechaFin)
	}
}

func TestHolguraCeroComportaNormal(t *testing.T) {
	turnosDef := []turnos.Turno{
		newTurno("DAY", "06:00:00", "18:00:00"),
		newTurno("NIGHT", "18:00:00", "06:00:00"),
	}

	input := TestSegment{
		FechaInicio: makeDate("2025-01-01 17:00"),
		FechaFin:    makeDate("2025-01-01 18:03"),
	}

	res := turnos.SegmentarPorTurnos(turnosDef, input, 0)

	esperados := []TestSegment{
		{"DAY", makeDate("2025-01-01 17:00"), makeDate("2025-01-01 18:00")},
		{"NIGHT", makeDate("2025-01-01 18:00"), makeDate("2025-01-01 18:03")},
	}

	require.Len(t, res, 2)

	for i := range esperados {
		require.Equal(t, esperados[i].CodigoTurno, res[i].CodigoTurno)
		require.Equal(t, esperados[i].FechaInicio, res[i].FechaInicio)
		require.Equal(t, esperados[i].FechaFin, res[i].FechaFin)
	}
}
