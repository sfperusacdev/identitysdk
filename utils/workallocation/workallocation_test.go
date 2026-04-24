package workallocation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/user0608/goones/types"
)

type testRange struct {
	id    string
	start time.Time
	end   time.Time
}

func (r testRange) StartTime() time.Time {
	return r.start
}

func (r testRange) EndTime() time.Time {
	return r.end
}

type testTurno struct {
	id    string
	start types.JustTime
	end   types.JustTime
}

func (t testTurno) TurnoID() string {
	return t.id
}

func (t testTurno) StartTurnoTime() types.JustTime {
	return t.start
}

func (t testTurno) EndTurnoTime() types.JustTime {
	return t.end
}

func dt(day int, hour int, minute int) time.Time {
	return time.Date(2026, 4, day, hour, minute, 0, 0, time.FixedZone("PET", -5*60*60))
}

func jt(hour int, minute int) types.JustTime {
	return types.JustTime(
		time.Duration(hour)*time.Hour +
			time.Duration(minute)*time.Minute,
	)
}

func defaultConfig() Config {
	return Config{
		RegularLimit: 8 * time.Hour,
		Extra25Limit: 2 * time.Hour,
	}
}

func defaultTurnos() []testTurno {
	return []testTurno{
		{id: "dia", start: jt(6, 0), end: jt(14, 0)},
		{id: "tarde", start: jt(14, 0), end: jt(22, 0)},
		{id: "noche", start: jt(22, 0), end: jt(6, 0)},
	}
}

func findTurnoAllocation(
	t *testing.T,
	allocation Allocation[testRange, testTurno],
	turnoID string,
) TurnoAllocation[testTurno] {
	t.Helper()

	for _, turnoAllocation := range allocation.Turnos {
		if turnoAllocation.TurnoID == turnoID {
			return turnoAllocation
		}
	}

	t.Fatalf("expected turno allocation %q to exist", turnoID)
	return TurnoAllocation[testTurno]{}
}

func assertNoTurnoAllocation(
	t *testing.T,
	allocation Allocation[testRange, testTurno],
	turnoID string,
) {
	t.Helper()

	for _, turnoAllocation := range allocation.Turnos {
		if turnoAllocation.TurnoID == turnoID {
			t.Fatalf("expected turno allocation %q to not exist", turnoID)
		}
	}
}

func TestServiceCalculate_DiscountsRestsClassifiesOvertimeAndAllocatesByTurno(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 8, 0), end: dt(24, 11, 0)},
		{id: "B", start: dt(24, 11, 0), end: dt(24, 15, 0)},
		{id: "C", start: dt(24, 15, 0), end: dt(24, 18, 0)},
		{id: "D", start: dt(24, 18, 0), end: dt(24, 22, 0)},
	}

	restRanges := []testRange{
		{id: "R1", start: dt(24, 12, 0), end: dt(24, 12, 30)},
		{id: "R2", start: dt(24, 19, 0), end: dt(24, 19, 30)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, restRanges, defaultTurnos())

	require.Len(t, got, 4)

	require.Equal(t, "A", got[0].Item.id)
	require.Equal(t, 3*time.Hour, got[0].Horas)
	require.Equal(t, 3*time.Hour, got[0].Regular)
	require.Equal(t, time.Duration(0), got[0].Extra25)
	require.Equal(t, time.Duration(0), got[0].Extra35)
	require.Equal(t, time.Duration(0), got[0].Descanso)

	aDia := findTurnoAllocation(t, got[0], "dia")
	require.Equal(t, 3*time.Hour, aDia.Horas)
	require.Equal(t, time.Duration(0), aDia.Extra25)
	require.Equal(t, time.Duration(0), aDia.Extra35)
	require.Len(t, aDia.Segments, 1)
	require.Equal(t, dt(24, 8, 0), aDia.Segments[0].Start)
	require.Equal(t, dt(24, 11, 0), aDia.Segments[0].End)
	require.Equal(t, Regular, aDia.Segments[0].Kind)

	require.Equal(t, "B", got[1].Item.id)
	require.Equal(t, 3*time.Hour+30*time.Minute, got[1].Horas)
	require.Equal(t, 3*time.Hour+30*time.Minute, got[1].Regular)
	require.Equal(t, time.Duration(0), got[1].Extra25)
	require.Equal(t, time.Duration(0), got[1].Extra35)
	require.Equal(t, 30*time.Minute, got[1].Descanso)

	bDia := findTurnoAllocation(t, got[1], "dia")
	require.Equal(t, 2*time.Hour+30*time.Minute, bDia.Horas)
	require.Equal(t, time.Duration(0), bDia.Extra25)
	require.Equal(t, time.Duration(0), bDia.Extra35)
	require.Len(t, bDia.Segments, 2)
	require.Equal(t, dt(24, 11, 0), bDia.Segments[0].Start)
	require.Equal(t, dt(24, 12, 0), bDia.Segments[0].End)
	require.Equal(t, Regular, bDia.Segments[0].Kind)
	require.Equal(t, dt(24, 12, 30), bDia.Segments[1].Start)
	require.Equal(t, dt(24, 14, 0), bDia.Segments[1].End)
	require.Equal(t, Regular, bDia.Segments[1].Kind)

	bTarde := findTurnoAllocation(t, got[1], "tarde")
	require.Equal(t, time.Hour, bTarde.Horas)
	require.Equal(t, time.Duration(0), bTarde.Extra25)
	require.Equal(t, time.Duration(0), bTarde.Extra35)
	require.Len(t, bTarde.Segments, 1)
	require.Equal(t, dt(24, 14, 0), bTarde.Segments[0].Start)
	require.Equal(t, dt(24, 15, 0), bTarde.Segments[0].End)
	require.Equal(t, Regular, bTarde.Segments[0].Kind)

	require.Equal(t, "C", got[2].Item.id)
	require.Equal(t, 3*time.Hour, got[2].Horas)
	require.Equal(t, 90*time.Minute, got[2].Regular)
	require.Equal(t, 90*time.Minute, got[2].Extra25)
	require.Equal(t, time.Duration(0), got[2].Extra35)
	require.Equal(t, time.Duration(0), got[2].Descanso)

	cTarde := findTurnoAllocation(t, got[2], "tarde")
	require.Equal(t, 3*time.Hour, cTarde.Horas)
	require.Equal(t, 90*time.Minute, cTarde.Extra25)
	require.Equal(t, time.Duration(0), cTarde.Extra35)
	require.Len(t, cTarde.Segments, 2)
	require.Equal(t, dt(24, 15, 0), cTarde.Segments[0].Start)
	require.Equal(t, dt(24, 16, 30), cTarde.Segments[0].End)
	require.Equal(t, Regular, cTarde.Segments[0].Kind)
	require.Equal(t, dt(24, 16, 30), cTarde.Segments[1].Start)
	require.Equal(t, dt(24, 18, 0), cTarde.Segments[1].End)
	require.Equal(t, Extra25, cTarde.Segments[1].Kind)

	require.Equal(t, "D", got[3].Item.id)
	require.Equal(t, 3*time.Hour+30*time.Minute, got[3].Horas)
	require.Equal(t, time.Duration(0), got[3].Regular)
	require.Equal(t, 30*time.Minute, got[3].Extra25)
	require.Equal(t, 3*time.Hour, got[3].Extra35)
	require.Equal(t, 30*time.Minute, got[3].Descanso)

	dTarde := findTurnoAllocation(t, got[3], "tarde")
	require.Equal(t, 3*time.Hour+30*time.Minute, dTarde.Horas)
	require.Equal(t, 30*time.Minute, dTarde.Extra25)
	require.Equal(t, 3*time.Hour, dTarde.Extra35)
	require.Len(t, dTarde.Segments, 3)
	require.Equal(t, dt(24, 18, 0), dTarde.Segments[0].Start)
	require.Equal(t, dt(24, 18, 30), dTarde.Segments[0].End)
	require.Equal(t, Extra25, dTarde.Segments[0].Kind)
	require.Equal(t, dt(24, 18, 30), dTarde.Segments[1].Start)
	require.Equal(t, dt(24, 19, 0), dTarde.Segments[1].End)
	require.Equal(t, Extra35, dTarde.Segments[1].Kind)
	require.Equal(t, dt(24, 19, 30), dTarde.Segments[2].Start)
	require.Equal(t, dt(24, 22, 0), dTarde.Segments[2].End)
	require.Equal(t, Extra35, dTarde.Segments[2].Kind)

	totalHoras := got[0].Horas + got[1].Horas + got[2].Horas + got[3].Horas
	totalRegular := got[0].Regular + got[1].Regular + got[2].Regular + got[3].Regular
	totalExtra25 := got[0].Extra25 + got[1].Extra25 + got[2].Extra25 + got[3].Extra25
	totalExtra35 := got[0].Extra35 + got[1].Extra35 + got[2].Extra35 + got[3].Extra35
	totalDescanso := got[0].Descanso + got[1].Descanso + got[2].Descanso + got[3].Descanso

	require.Equal(t, 13*time.Hour, totalHoras)
	require.Equal(t, 8*time.Hour, totalRegular)
	require.Equal(t, 2*time.Hour, totalExtra25)
	require.Equal(t, 3*time.Hour, totalExtra35)
	require.Equal(t, time.Hour, totalDescanso)
}

func TestServiceCalculate_WorkRangeFullyCoveredByRestStillReturnsAllocation(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 8, 0), end: dt(24, 10, 0)},
	}

	restRanges := []testRange{
		{id: "R1", start: dt(24, 7, 0), end: dt(24, 11, 0)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, restRanges, defaultTurnos())

	require.Len(t, got, 1)
	require.Equal(t, "A", got[0].Item.id)
	require.Equal(t, time.Duration(0), got[0].Horas)
	require.Equal(t, time.Duration(0), got[0].Regular)
	require.Equal(t, time.Duration(0), got[0].Extra25)
	require.Equal(t, time.Duration(0), got[0].Extra35)
	require.Equal(t, 2*time.Hour, got[0].Descanso)
	require.Len(t, got[0].Descansos, 1)
	require.Equal(t, dt(24, 8, 0), got[0].Descansos[0].Start)
	require.Equal(t, dt(24, 10, 0), got[0].Descansos[0].End)
	require.Empty(t, got[0].Turnos)
}

func TestServiceCalculate_AllocatesWorkCrossingMidnightToNightTurno(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 21, 0), end: dt(25, 2, 0)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, nil, defaultTurnos())

	require.Len(t, got, 1)
	require.Equal(t, 5*time.Hour, got[0].Horas)
	require.Equal(t, 5*time.Hour, got[0].Regular)
	require.Equal(t, time.Duration(0), got[0].Extra25)
	require.Equal(t, time.Duration(0), got[0].Extra35)

	tarde := findTurnoAllocation(t, got[0], "tarde")
	require.Equal(t, time.Hour, tarde.Horas)
	require.Equal(t, time.Duration(0), tarde.Extra25)
	require.Equal(t, time.Duration(0), tarde.Extra35)
	require.Len(t, tarde.Segments, 1)
	require.Equal(t, dt(24, 21, 0), tarde.Segments[0].Start)
	require.Equal(t, dt(24, 22, 0), tarde.Segments[0].End)
	require.Equal(t, Regular, tarde.Segments[0].Kind)

	noche := findTurnoAllocation(t, got[0], "noche")
	require.Equal(t, 4*time.Hour, noche.Horas)
	require.Equal(t, time.Duration(0), noche.Extra25)
	require.Equal(t, time.Duration(0), noche.Extra35)
	require.Len(t, noche.Segments, 1)
	require.Equal(t, dt(24, 22, 0), noche.Segments[0].Start)
	require.Equal(t, dt(25, 2, 0), noche.Segments[0].End)
	require.Equal(t, Regular, noche.Segments[0].Kind)

	assertNoTurnoAllocation(t, got[0], "dia")
}

func TestServiceCalculate_OvertimeCrossesTurnoBoundary(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 8, 0), end: dt(24, 17, 0)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, nil, defaultTurnos())

	require.Len(t, got, 1)
	require.Equal(t, 9*time.Hour, got[0].Horas)
	require.Equal(t, 8*time.Hour, got[0].Regular)
	require.Equal(t, time.Hour, got[0].Extra25)
	require.Equal(t, time.Duration(0), got[0].Extra35)

	dia := findTurnoAllocation(t, got[0], "dia")
	require.Equal(t, 6*time.Hour, dia.Horas)
	require.Equal(t, time.Duration(0), dia.Extra25)
	require.Equal(t, time.Duration(0), dia.Extra35)

	tarde := findTurnoAllocation(t, got[0], "tarde")
	require.Equal(t, 3*time.Hour, tarde.Horas)
	require.Equal(t, time.Hour, tarde.Extra25)
	require.Equal(t, time.Duration(0), tarde.Extra35)

	require.Len(t, tarde.Segments, 2)
	require.Equal(t, dt(24, 14, 0), tarde.Segments[0].Start)
	require.Equal(t, dt(24, 16, 0), tarde.Segments[0].End)
	require.Equal(t, Regular, tarde.Segments[0].Kind)
	require.Equal(t, dt(24, 16, 0), tarde.Segments[1].Start)
	require.Equal(t, dt(24, 17, 0), tarde.Segments[1].End)
	require.Equal(t, Extra25, tarde.Segments[1].Kind)
}

func TestServiceCalculate_Extra35CrossesMidnightTurno(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 12, 0), end: dt(25, 1, 0)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, nil, defaultTurnos())

	require.Len(t, got, 1)
	require.Equal(t, 13*time.Hour, got[0].Horas)
	require.Equal(t, 8*time.Hour, got[0].Regular)
	require.Equal(t, 2*time.Hour, got[0].Extra25)
	require.Equal(t, 3*time.Hour, got[0].Extra35)

	dia := findTurnoAllocation(t, got[0], "dia")
	require.Equal(t, 2*time.Hour, dia.Horas)
	require.Equal(t, time.Duration(0), dia.Extra25)
	require.Equal(t, time.Duration(0), dia.Extra35)

	tarde := findTurnoAllocation(t, got[0], "tarde")
	require.Equal(t, 8*time.Hour, tarde.Horas)
	require.Equal(t, 2*time.Hour, tarde.Extra25)
	require.Equal(t, time.Duration(0), tarde.Extra35)

	noche := findTurnoAllocation(t, got[0], "noche")
	require.Equal(t, 3*time.Hour, noche.Horas)
	require.Equal(t, time.Duration(0), noche.Extra25)
	require.Equal(t, 3*time.Hour, noche.Extra35)

	require.Len(t, noche.Segments, 1)
	require.Equal(t, dt(24, 22, 0), noche.Segments[0].Start)
	require.Equal(t, dt(25, 1, 0), noche.Segments[0].End)
	require.Equal(t, Extra35, noche.Segments[0].Kind)
}

func TestServiceCalculate_EmptyInput(t *testing.T) {
	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(nil, nil, defaultTurnos())

	require.Empty(t, got)
}

func TestServiceCalculate_ClassicOvertimeExampleWithSingleFullDayTurno(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 8, 0), end: dt(24, 11, 0)},
		{id: "B", start: dt(24, 11, 0), end: dt(24, 15, 0)},
		{id: "C", start: dt(24, 15, 0), end: dt(24, 18, 0)},
		{id: "D", start: dt(24, 18, 0), end: dt(24, 22, 0)},
	}

	turnos := []testTurno{
		{id: "full-day", start: jt(0, 0), end: jt(0, 0)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, nil, turnos)

	require.Len(t, got, 4)

	require.Equal(t, "A", got[0].Item.id)
	require.Equal(t, 3*time.Hour, got[0].Horas)
	require.Equal(t, 3*time.Hour, got[0].Regular)
	require.Equal(t, time.Duration(0), got[0].Extra25)
	require.Equal(t, time.Duration(0), got[0].Extra35)

	require.Equal(t, "B", got[1].Item.id)
	require.Equal(t, 4*time.Hour, got[1].Horas)
	require.Equal(t, 4*time.Hour, got[1].Regular)
	require.Equal(t, time.Duration(0), got[1].Extra25)
	require.Equal(t, time.Duration(0), got[1].Extra35)

	require.Equal(t, "C", got[2].Item.id)
	require.Equal(t, 3*time.Hour, got[2].Horas)
	require.Equal(t, time.Hour, got[2].Regular)
	require.Equal(t, 2*time.Hour, got[2].Extra25)
	require.Equal(t, time.Duration(0), got[2].Extra35)

	require.Equal(t, "D", got[3].Item.id)
	require.Equal(t, 4*time.Hour, got[3].Horas)
	require.Equal(t, time.Duration(0), got[3].Regular)
	require.Equal(t, time.Duration(0), got[3].Extra25)
	require.Equal(t, 4*time.Hour, got[3].Extra35)

	totalHoras := got[0].Horas + got[1].Horas + got[2].Horas + got[3].Horas
	totalRegular := got[0].Regular + got[1].Regular + got[2].Regular + got[3].Regular
	totalExtra25 := got[0].Extra25 + got[1].Extra25 + got[2].Extra25 + got[3].Extra25
	totalExtra35 := got[0].Extra35 + got[1].Extra35 + got[2].Extra35 + got[3].Extra35

	require.Equal(t, 14*time.Hour, totalHoras)
	require.Equal(t, 8*time.Hour, totalRegular)
	require.Equal(t, 2*time.Hour, totalExtra25)
	require.Equal(t, 4*time.Hour, totalExtra35)

	for _, allocation := range got {
		fullDay := findTurnoAllocation(t, allocation, "full-day")
		require.Equal(t, allocation.Horas, fullDay.Horas)
		require.Equal(t, allocation.Extra25, fullDay.Extra25)
		require.Equal(t, allocation.Extra35, fullDay.Extra35)
	}
}

func TestServiceCalculate_TracksDescansoTotalsAndSegments(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 8, 0), end: dt(24, 12, 0)},
		{id: "B", start: dt(24, 13, 0), end: dt(24, 18, 0)},
	}

	restRanges := []testRange{
		{id: "R1", start: dt(24, 9, 0), end: dt(24, 9, 30)},
		{id: "R2", start: dt(24, 11, 30), end: dt(24, 13, 30)},
		{id: "R3", start: dt(24, 17, 0), end: dt(24, 18, 30)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, restRanges, defaultTurnos())

	require.Len(t, got, 2)

	require.Equal(t, "A", got[0].Item.id)
	require.Equal(t, 3*time.Hour, got[0].Horas)
	require.Equal(t, time.Hour, got[0].Descanso)
	require.Len(t, got[0].Descansos, 2)

	require.Equal(t, dt(24, 9, 0), got[0].Descansos[0].Start)
	require.Equal(t, dt(24, 9, 30), got[0].Descansos[0].End)

	require.Equal(t, dt(24, 11, 30), got[0].Descansos[1].Start)
	require.Equal(t, dt(24, 12, 0), got[0].Descansos[1].End)

	require.Equal(t, 3*time.Hour, got[0].Regular)
	require.Equal(t, time.Duration(0), got[0].Extra25)
	require.Equal(t, time.Duration(0), got[0].Extra35)

	require.Equal(t, "B", got[1].Item.id)
	require.Equal(t, 3*time.Hour+30*time.Minute, got[1].Horas)
	require.Equal(t, 90*time.Minute, got[1].Descanso)
	require.Len(t, got[1].Descansos, 2)

	require.Equal(t, dt(24, 13, 0), got[1].Descansos[0].Start)
	require.Equal(t, dt(24, 13, 30), got[1].Descansos[0].End)

	require.Equal(t, dt(24, 17, 0), got[1].Descansos[1].Start)
	require.Equal(t, dt(24, 18, 0), got[1].Descansos[1].End)

	require.Equal(t, 3*time.Hour+30*time.Minute, got[1].Regular)
	require.Equal(t, time.Duration(0), got[1].Extra25)
	require.Equal(t, time.Duration(0), got[1].Extra35)

	totalDescanso := got[0].Descanso + got[1].Descanso
	totalWorked := got[0].Horas + got[1].Horas

	require.Equal(t, 150*time.Minute, totalDescanso)
	require.Equal(t, 6*time.Hour+30*time.Minute, totalWorked)
}

func TestServiceCalculate_NightShiftWithRefrigerioCrossingMidnight(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 17, 0), end: dt(24, 20, 0)},
		{id: "B", start: dt(24, 20, 0), end: dt(25, 0, 0)},
		{id: "C", start: dt(25, 0, 0), end: dt(25, 3, 0)},
		{id: "D", start: dt(25, 3, 0), end: dt(25, 8, 0)},
	}

	restRanges := []testRange{
		{id: "REFRIGERIO", start: dt(24, 23, 30), end: dt(25, 0, 30)},
	}

	turnos := []testTurno{
		{id: "DIURNO", start: jt(6, 0), end: jt(22, 0)},
		{id: "NOCTURNO", start: jt(22, 0), end: jt(6, 0)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, restRanges, turnos)

	require.Len(t, got, 4)

	require.Equal(t, "A", got[0].Item.id)
	require.Equal(t, 3*time.Hour, got[0].Horas)
	require.Equal(t, 3*time.Hour, got[0].Regular)
	require.Equal(t, time.Duration(0), got[0].Extra25)
	require.Equal(t, time.Duration(0), got[0].Extra35)
	require.Equal(t, time.Duration(0), got[0].Descanso)
	require.Empty(t, got[0].Descansos)

	aDiurno := findTurnoAllocation(t, got[0], "DIURNO")
	require.Equal(t, 3*time.Hour, aDiurno.Horas)
	require.Equal(t, time.Duration(0), aDiurno.Extra25)
	require.Equal(t, time.Duration(0), aDiurno.Extra35)

	require.Equal(t, "B", got[1].Item.id)
	require.Equal(t, 3*time.Hour+30*time.Minute, got[1].Horas)
	require.Equal(t, 3*time.Hour+30*time.Minute, got[1].Regular)
	require.Equal(t, time.Duration(0), got[1].Extra25)
	require.Equal(t, time.Duration(0), got[1].Extra35)
	require.Equal(t, 30*time.Minute, got[1].Descanso)
	require.Len(t, got[1].Descansos, 1)
	require.Equal(t, dt(24, 23, 30), got[1].Descansos[0].Start)
	require.Equal(t, dt(25, 0, 0), got[1].Descansos[0].End)

	bDiurno := findTurnoAllocation(t, got[1], "DIURNO")
	require.Equal(t, 2*time.Hour, bDiurno.Horas)
	require.Equal(t, time.Duration(0), bDiurno.Extra25)
	require.Equal(t, time.Duration(0), bDiurno.Extra35)

	bNocturno := findTurnoAllocation(t, got[1], "NOCTURNO")
	require.Equal(t, 90*time.Minute, bNocturno.Horas)
	require.Equal(t, time.Duration(0), bNocturno.Extra25)
	require.Equal(t, time.Duration(0), bNocturno.Extra35)

	require.Equal(t, "C", got[2].Item.id)
	require.Equal(t, 2*time.Hour+30*time.Minute, got[2].Horas)
	require.Equal(t, 90*time.Minute, got[2].Regular)
	require.Equal(t, time.Hour, got[2].Extra25)
	require.Equal(t, time.Duration(0), got[2].Extra35)
	require.Equal(t, 30*time.Minute, got[2].Descanso)
	require.Len(t, got[2].Descansos, 1)
	require.Equal(t, dt(25, 0, 0), got[2].Descansos[0].Start)
	require.Equal(t, dt(25, 0, 30), got[2].Descansos[0].End)

	cNocturno := findTurnoAllocation(t, got[2], "NOCTURNO")
	require.Equal(t, 2*time.Hour+30*time.Minute, cNocturno.Horas)
	require.Equal(t, time.Hour, cNocturno.Extra25)
	require.Equal(t, time.Duration(0), cNocturno.Extra35)

	require.Equal(t, "D", got[3].Item.id)
	require.Equal(t, 5*time.Hour, got[3].Horas)
	require.Equal(t, time.Duration(0), got[3].Regular)
	require.Equal(t, time.Hour, got[3].Extra25)
	require.Equal(t, 4*time.Hour, got[3].Extra35)
	require.Equal(t, time.Duration(0), got[3].Descanso)
	require.Empty(t, got[3].Descansos)

	dNocturno := findTurnoAllocation(t, got[3], "NOCTURNO")
	require.Equal(t, 3*time.Hour, dNocturno.Horas)
	require.Equal(t, time.Hour, dNocturno.Extra25)
	require.Equal(t, 2*time.Hour, dNocturno.Extra35)

	dDiurno := findTurnoAllocation(t, got[3], "DIURNO")
	require.Equal(t, 2*time.Hour, dDiurno.Horas)
	require.Equal(t, time.Duration(0), dDiurno.Extra25)
	require.Equal(t, 2*time.Hour, dDiurno.Extra35)

	totalHoras := got[0].Horas + got[1].Horas + got[2].Horas + got[3].Horas
	totalRegular := got[0].Regular + got[1].Regular + got[2].Regular + got[3].Regular
	totalExtra25 := got[0].Extra25 + got[1].Extra25 + got[2].Extra25 + got[3].Extra25
	totalExtra35 := got[0].Extra35 + got[1].Extra35 + got[2].Extra35 + got[3].Extra35
	totalDescanso := got[0].Descanso + got[1].Descanso + got[2].Descanso + got[3].Descanso

	require.Equal(t, 14*time.Hour, totalHoras)
	require.Equal(t, 8*time.Hour, totalRegular)
	require.Equal(t, 2*time.Hour, totalExtra25)
	require.Equal(t, 4*time.Hour, totalExtra35)
	require.Equal(t, time.Hour, totalDescanso)
}

func TestServiceCalculate_GapsBetweenWorkRanges_Simple(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 8, 0), end: dt(24, 10, 0)},
		{id: "B", start: dt(24, 10, 10), end: dt(24, 12, 10)},
		{id: "C", start: dt(24, 12, 30), end: dt(24, 15, 30)},
		{id: "D", start: dt(24, 16, 0), end: dt(24, 19, 0)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, nil, defaultTurnos())

	require.Len(t, got, 4)

	require.Equal(t, 2*time.Hour, got[0].Regular)
	require.Equal(t, 2*time.Hour, got[1].Regular)
	require.Equal(t, 3*time.Hour, got[2].Regular)

	require.Equal(t, time.Hour, got[3].Regular)
	require.Equal(t, 2*time.Hour, got[3].Extra25)

	totalRegular := got[0].Regular + got[1].Regular + got[2].Regular + got[3].Regular
	totalExtra25 := got[3].Extra25

	require.Equal(t, 8*time.Hour, totalRegular)
	require.Equal(t, 2*time.Hour, totalExtra25)
}

func TestServiceCalculate_GapsWithRestInside(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 8, 0), end: dt(24, 10, 0)},
		{id: "B", start: dt(24, 10, 15), end: dt(24, 13, 0)},
		{id: "C", start: dt(24, 13, 45), end: dt(24, 17, 15)},
		{id: "D", start: dt(24, 17, 30), end: dt(24, 20, 30)},
	}

	restRanges := []testRange{
		{id: "R1", start: dt(24, 12, 0), end: dt(24, 12, 30)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, restRanges, defaultTurnos())

	require.Len(t, got, 4)

	require.Equal(t, 2*time.Hour, got[0].Regular)

	require.Equal(t, 2*time.Hour+15*time.Minute, got[1].Regular)
	require.Equal(t, 30*time.Minute, got[1].Descanso)

	require.Equal(t, 3*time.Hour+30*time.Minute, got[2].Regular)

	require.Equal(t, 15*time.Minute, got[3].Regular)
	require.Equal(t, 2*time.Hour, got[3].Extra25)
	require.Equal(t, 45*time.Minute, got[3].Extra35)

	totalDescanso := got[1].Descanso
	require.Equal(t, 30*time.Minute, totalDescanso)
}

func TestServiceCalculate_GapsCrossingNightShift(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 18, 0), end: dt(24, 21, 0)},
		{id: "B", start: dt(24, 21, 20), end: dt(24, 23, 40)},
		{id: "C", start: dt(25, 0, 10), end: dt(25, 3, 10)},
		{id: "D", start: dt(25, 3, 30), end: dt(25, 7, 30)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, nil, defaultTurnos())

	require.Len(t, got, 4)

	require.Equal(t, 3*time.Hour, got[0].Regular)

	require.Equal(t, 2*time.Hour+20*time.Minute, got[1].Regular)

	require.Equal(t, 2*time.Hour+40*time.Minute, got[2].Regular)
	require.Equal(t, 20*time.Minute, got[2].Extra25)

	require.Equal(t, time.Duration(0), got[3].Regular)
	require.Equal(t, 1*time.Hour+40*time.Minute, got[3].Extra25)
	require.Equal(t, 2*time.Hour+20*time.Minute, got[3].Extra35)
}

func TestServiceCalculate_GapsComplexScenarioWithCrossMidnightRest(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 17, 0), end: dt(24, 19, 0)},
		{id: "B", start: dt(24, 19, 20), end: dt(24, 22, 20)},
		{id: "C", start: dt(24, 22, 40), end: dt(25, 1, 20)},
		{id: "D", start: dt(25, 1, 45), end: dt(25, 4, 45)},
		{id: "E", start: dt(25, 5, 0), end: dt(25, 7, 0)},
	}

	restRanges := []testRange{
		{id: "R1", start: dt(24, 23, 30), end: dt(25, 0, 30)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, restRanges, defaultTurnos())

	require.Len(t, got, 5)

	require.Equal(t, 2*time.Hour, got[0].Regular)
	require.Equal(t, 3*time.Hour, got[1].Regular)

	require.Equal(t, 1*time.Hour+40*time.Minute, got[2].Regular)
	require.Equal(t, time.Hour, got[2].Descanso)

	require.Equal(t, 1*time.Hour+20*time.Minute, got[3].Regular)
	require.Equal(t, 1*time.Hour+40*time.Minute, got[3].Extra25)

	require.Equal(t, 20*time.Minute, got[4].Extra25)
	require.Equal(t, 1*time.Hour+40*time.Minute, got[4].Extra35)

	totalRegular := time.Duration(0)
	totalExtra25 := time.Duration(0)
	totalExtra35 := time.Duration(0)
	totalDescanso := time.Duration(0)

	for _, a := range got {
		totalRegular += a.Regular
		totalExtra25 += a.Extra25
		totalExtra35 += a.Extra35
		totalDescanso += a.Descanso
	}

	require.Equal(t, 8*time.Hour, totalRegular)
	require.Equal(t, 2*time.Hour, totalExtra25)
	require.Equal(t, 1*time.Hour+40*time.Minute, totalExtra35)
	require.Equal(t, time.Hour, totalDescanso)
}

func TestServiceCalculate_FullNightCaseWithRefrigerioAndTurnos(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 17, 0), end: dt(24, 20, 0)},
		{id: "B", start: dt(24, 20, 0), end: dt(25, 0, 0)},
		{id: "C", start: dt(25, 0, 0), end: dt(25, 3, 0)},
		{id: "D", start: dt(25, 3, 0), end: dt(25, 8, 0)},
	}

	restRanges := []testRange{
		{id: "REFRIGERIO", start: dt(24, 23, 30), end: dt(25, 0, 30)},
	}

	turnos := []testTurno{
		{id: "DIURNO", start: jt(6, 0), end: jt(22, 0)},
		{id: "NOCTURNO", start: jt(22, 0), end: jt(6, 0)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, restRanges, turnos)

	require.Len(t, got, 4)

	require.Equal(t, "A", got[0].Item.id)
	require.Equal(t, 3*time.Hour, got[0].Horas)
	require.Equal(t, 3*time.Hour, got[0].Regular)
	require.Equal(t, time.Duration(0), got[0].Extra25)
	require.Equal(t, time.Duration(0), got[0].Extra35)
	require.Equal(t, time.Duration(0), got[0].Descanso)
	require.Empty(t, got[0].Descansos)

	aDiurno := findTurnoAllocation(t, got[0], "DIURNO")
	require.Equal(t, 3*time.Hour, aDiurno.Horas)
	require.Equal(t, time.Duration(0), aDiurno.Extra25)
	require.Equal(t, time.Duration(0), aDiurno.Extra35)

	require.Equal(t, "B", got[1].Item.id)
	require.Equal(t, 3*time.Hour+30*time.Minute, got[1].Horas)
	require.Equal(t, 3*time.Hour+30*time.Minute, got[1].Regular)
	require.Equal(t, time.Duration(0), got[1].Extra25)
	require.Equal(t, time.Duration(0), got[1].Extra35)
	require.Equal(t, 30*time.Minute, got[1].Descanso)
	require.Len(t, got[1].Descansos, 1)
	require.Equal(t, dt(24, 23, 30), got[1].Descansos[0].Start)
	require.Equal(t, dt(25, 0, 0), got[1].Descansos[0].End)

	bDiurno := findTurnoAllocation(t, got[1], "DIURNO")
	require.Equal(t, 2*time.Hour, bDiurno.Horas)
	require.Equal(t, time.Duration(0), bDiurno.Extra25)
	require.Equal(t, time.Duration(0), bDiurno.Extra35)

	bNocturno := findTurnoAllocation(t, got[1], "NOCTURNO")
	require.Equal(t, 90*time.Minute, bNocturno.Horas)
	require.Equal(t, time.Duration(0), bNocturno.Extra25)
	require.Equal(t, time.Duration(0), bNocturno.Extra35)

	require.Equal(t, "C", got[2].Item.id)
	require.Equal(t, 2*time.Hour+30*time.Minute, got[2].Horas)
	require.Equal(t, 90*time.Minute, got[2].Regular)
	require.Equal(t, time.Hour, got[2].Extra25)
	require.Equal(t, time.Duration(0), got[2].Extra35)
	require.Equal(t, 30*time.Minute, got[2].Descanso)
	require.Len(t, got[2].Descansos, 1)
	require.Equal(t, dt(25, 0, 0), got[2].Descansos[0].Start)
	require.Equal(t, dt(25, 0, 30), got[2].Descansos[0].End)

	cNocturno := findTurnoAllocation(t, got[2], "NOCTURNO")
	require.Equal(t, 2*time.Hour+30*time.Minute, cNocturno.Horas)
	require.Equal(t, time.Hour, cNocturno.Extra25)
	require.Equal(t, time.Duration(0), cNocturno.Extra35)

	require.Equal(t, "D", got[3].Item.id)
	require.Equal(t, 5*time.Hour, got[3].Horas)
	require.Equal(t, time.Duration(0), got[3].Regular)
	require.Equal(t, time.Hour, got[3].Extra25)
	require.Equal(t, 4*time.Hour, got[3].Extra35)
	require.Equal(t, time.Duration(0), got[3].Descanso)
	require.Empty(t, got[3].Descansos)

	dNocturno := findTurnoAllocation(t, got[3], "NOCTURNO")
	require.Equal(t, 3*time.Hour, dNocturno.Horas)
	require.Equal(t, time.Hour, dNocturno.Extra25)
	require.Equal(t, 2*time.Hour, dNocturno.Extra35)

	dDiurno := findTurnoAllocation(t, got[3], "DIURNO")
	require.Equal(t, 2*time.Hour, dDiurno.Horas)
	require.Equal(t, time.Duration(0), dDiurno.Extra25)
	require.Equal(t, 2*time.Hour, dDiurno.Extra35)

	totalHoras := got[0].Horas + got[1].Horas + got[2].Horas + got[3].Horas
	totalRegular := got[0].Regular + got[1].Regular + got[2].Regular + got[3].Regular
	totalExtra25 := got[0].Extra25 + got[1].Extra25 + got[2].Extra25 + got[3].Extra25
	totalExtra35 := got[0].Extra35 + got[1].Extra35 + got[2].Extra35 + got[3].Extra35
	totalDescanso := got[0].Descanso + got[1].Descanso + got[2].Descanso + got[3].Descanso

	require.Equal(t, 14*time.Hour, totalHoras)
	require.Equal(t, 8*time.Hour, totalRegular)
	require.Equal(t, 2*time.Hour, totalExtra25)
	require.Equal(t, 4*time.Hour, totalExtra35)
	require.Equal(t, time.Hour, totalDescanso)
}

func TestServiceCalculate_WorkFullyCoveredByRestAcrossMidnight(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 22, 0), end: dt(25, 2, 0)},
	}

	restRanges := []testRange{
		{id: "REST", start: dt(24, 21, 0), end: dt(25, 3, 0)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, restRanges, defaultTurnos())

	require.Len(t, got, 1)
	require.Equal(t, time.Duration(0), got[0].Horas)
	require.Equal(t, time.Duration(0), got[0].Regular)
	require.Equal(t, time.Duration(0), got[0].Extra25)
	require.Equal(t, time.Duration(0), got[0].Extra35)
	require.Equal(t, 4*time.Hour, got[0].Descanso)
	require.Len(t, got[0].Descansos, 1)
	require.Equal(t, dt(24, 22, 0), got[0].Descansos[0].Start)
	require.Equal(t, dt(25, 2, 0), got[0].Descansos[0].End)
	require.Empty(t, got[0].Turnos)
}

func TestServiceCalculate_SingleLongWorkCrossesAllBandsAndTurnos(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 5, 0), end: dt(24, 23, 0)},
	}

	turnos := []testTurno{
		{id: "DIURNO", start: jt(6, 0), end: jt(22, 0)},
		{id: "NOCTURNO", start: jt(22, 0), end: jt(6, 0)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, nil, turnos)

	require.Len(t, got, 1)
	require.Equal(t, 18*time.Hour, got[0].Horas)
	require.Equal(t, 8*time.Hour, got[0].Regular)
	require.Equal(t, 2*time.Hour, got[0].Extra25)
	require.Equal(t, 8*time.Hour, got[0].Extra35)

	nocturno := findTurnoAllocation(t, got[0], "NOCTURNO")
	require.Equal(t, 2*time.Hour, nocturno.Horas)
	require.Equal(t, time.Duration(0), nocturno.Extra25)
	require.Equal(t, time.Hour, nocturno.Extra35)

	diurno := findTurnoAllocation(t, got[0], "DIURNO")
	require.Equal(t, 16*time.Hour, diurno.Horas)
	require.Equal(t, 2*time.Hour, diurno.Extra25)
	require.Equal(t, 7*time.Hour, diurno.Extra35)
}

func TestServiceCalculate_RestSplitsExtraBand(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 8, 0), end: dt(24, 19, 0)},
	}

	restRanges := []testRange{
		{id: "REST", start: dt(24, 16, 30), end: dt(24, 17, 30)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, restRanges, defaultTurnos())

	require.Len(t, got, 1)
	require.Equal(t, 10*time.Hour, got[0].Horas)
	require.Equal(t, 8*time.Hour, got[0].Regular)
	require.Equal(t, 2*time.Hour, got[0].Extra25)
	require.Equal(t, time.Duration(0), got[0].Extra35)
	require.Equal(t, time.Hour, got[0].Descanso)

	tarde := findTurnoAllocation(t, got[0], "tarde")
	require.Equal(t, 4*time.Hour, tarde.Horas)
	require.Equal(t, 2*time.Hour, tarde.Extra25)
	require.Equal(t, time.Duration(0), tarde.Extra35)

	require.Len(t, tarde.Segments, 3)
	require.Equal(t, dt(24, 14, 0), tarde.Segments[0].Start)
	require.Equal(t, dt(24, 16, 0), tarde.Segments[0].End)
	require.Equal(t, Regular, tarde.Segments[0].Kind)

	require.Equal(t, dt(24, 16, 0), tarde.Segments[1].Start)
	require.Equal(t, dt(24, 16, 30), tarde.Segments[1].End)
	require.Equal(t, Extra25, tarde.Segments[1].Kind)

	require.Equal(t, dt(24, 17, 30), tarde.Segments[2].Start)
	require.Equal(t, dt(24, 19, 0), tarde.Segments[2].End)
	require.Equal(t, Extra25, tarde.Segments[2].Kind)
}

func TestServiceCalculate_MultipleWorkRangesAndRestAcrossTurnoBoundary(t *testing.T) {
	workRanges := []testRange{
		{id: "A", start: dt(24, 13, 0), end: dt(24, 18, 0)},
		{id: "B", start: dt(24, 18, 0), end: dt(25, 3, 0)},
	}

	restRanges := []testRange{
		{id: "REST", start: dt(24, 21, 30), end: dt(24, 22, 30)},
	}

	turnos := []testTurno{
		{id: "DIURNO", start: jt(6, 0), end: jt(22, 0)},
		{id: "NOCTURNO", start: jt(22, 0), end: jt(6, 0)},
	}

	svc := NewService[testRange, testRange, testTurno](defaultConfig())

	got := svc.Calculate(workRanges, restRanges, turnos)

	require.Len(t, got, 2)

	require.Equal(t, "A", got[0].Item.id)
	require.Equal(t, 5*time.Hour, got[0].Horas)
	require.Equal(t, 5*time.Hour, got[0].Regular)
	require.Equal(t, time.Duration(0), got[0].Extra25)
	require.Equal(t, time.Duration(0), got[0].Extra35)

	aDiurno := findTurnoAllocation(t, got[0], "DIURNO")
	require.Equal(t, 5*time.Hour, aDiurno.Horas)

	require.Equal(t, "B", got[1].Item.id)
	require.Equal(t, 8*time.Hour, got[1].Horas)
	require.Equal(t, 3*time.Hour, got[1].Regular)
	require.Equal(t, 2*time.Hour, got[1].Extra25)
	require.Equal(t, 3*time.Hour, got[1].Extra35)
	require.Equal(t, time.Hour, got[1].Descanso)

	bDiurno := findTurnoAllocation(t, got[1], "DIURNO")
	require.Equal(t, 3*time.Hour+30*time.Minute, bDiurno.Horas)
	require.Equal(t, 30*time.Minute, bDiurno.Extra25)
	require.Equal(t, time.Duration(0), bDiurno.Extra35)

	bNocturno := findTurnoAllocation(t, got[1], "NOCTURNO")
	require.Equal(t, 4*time.Hour+30*time.Minute, bNocturno.Horas)
	require.Equal(t, 90*time.Minute, bNocturno.Extra25)
	require.Equal(t, 3*time.Hour, bNocturno.Extra35)
}
