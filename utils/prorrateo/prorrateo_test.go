package prorrateo

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

type testEvent struct {
	t time.Time
	v decimal.Decimal
}

func (e testEvent) Time() time.Time {
	return e.t
}

func (e testEvent) Amount() decimal.Decimal {
	return e.v
}

type testEntity struct {
	start time.Time
	end   time.Time
}

func (e testEntity) StartTime() time.Time {
	return e.start
}

func (e testEntity) EndTime() time.Time {
	return e.end
}

// Valida que un evento que ocurre exactamente en el StartTime de la entidad
// sea incluido correctamente en el prorrateo.
func TestEventAtStartTimeIncluded(t *testing.T) {
	t0 := time.Now()

	entities := []testEntity{
		{start: t0, end: t0.Add(time.Hour)},
	}

	events := []testEvent{
		{t: t0, v: decimal.NewFromInt(100)},
	}

	result := AllocateOverTime(events, entities)

	if !result[0].Total.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("expected 100, got %s", result[0].Total)
	}
}

// Valida que un evento que ocurre exactamente en el EndTime de la entidad
// no sea considerado en el prorrateo (EndTime es exclusivo).
func TestEventAtEndTimeExcluded(t *testing.T) {
	t0 := time.Now()

	entities := []testEntity{
		{start: t0, end: t0.Add(time.Hour)},
	}

	events := []testEvent{
		{t: t0.Add(time.Hour), v: decimal.NewFromInt(100)},
	}

	result := AllocateOverTime(events, entities)

	if !result[0].Total.Equal(decimal.Zero) {
		t.Fatalf("expected 0, got %s", result[0].Total)
	}
}

// Valida que múltiples eventos que ocurren en el mismo instante
// se acumulen correctamente y se repartan de forma equitativa.
func TestMultipleEventsSameInstant(t *testing.T) {
	t0 := time.Now()

	entities := []testEntity{
		{start: t0, end: t0.Add(time.Hour)},
		{start: t0, end: t0.Add(time.Hour)},
	}

	events := []testEvent{
		{t: t0.Add(10 * time.Minute), v: decimal.NewFromInt(60)},
		{t: t0.Add(10 * time.Minute), v: decimal.NewFromInt(40)},
	}

	result := AllocateOverTime(events, entities)

	expected := decimal.NewFromInt(50)

	for i, r := range result {
		if !r.Total.Equal(expected) {
			t.Fatalf("entity %d expected %s, got %s", i, expected, r.Total)
		}
	}
}

// Valida que el algoritmo sea independiente del orden de los eventos
// y produzca el mismo resultado aunque estén desordenados.
func TestEventsOutOfOrder(t *testing.T) {
	t0 := time.Now()

	entities := []testEntity{
		{start: t0, end: t0.Add(time.Hour)},
	}

	events := []testEvent{
		{t: t0.Add(30 * time.Minute), v: decimal.NewFromInt(30)},
		{t: t0.Add(10 * time.Minute), v: decimal.NewFromInt(70)},
	}

	result := AllocateOverTime(events, entities)

	if !result[0].Total.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("expected 100, got %s", result[0].Total)
	}
}

// Valida que la división con decimales preserve la suma total
// cuando el valor no es divisible de forma exacta.
func TestDecimalPrecision(t *testing.T) {
	t0 := time.Now()

	entities := []testEntity{
		{start: t0, end: t0.Add(time.Hour)},
		{start: t0, end: t0.Add(time.Hour)},
		{start: t0, end: t0.Add(time.Hour)},
	}

	events := []testEvent{
		{t: t0.Add(10 * time.Minute), v: decimal.NewFromInt(100)},
	}

	result := AllocateOverTime(events, entities)

	sum := decimal.Zero
	for _, r := range result {
		sum = sum.Add(r.Total)
	}

	if !sum.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("expected sum 100, got %s", sum)
	}
}

// Valida que un evento con valor cero no afecte el resultado final.
func TestZeroValueEvent(t *testing.T) {
	t0 := time.Now()

	entities := []testEntity{
		{start: t0, end: t0.Add(time.Hour)},
	}

	events := []testEvent{
		{t: t0.Add(10 * time.Minute), v: decimal.Zero},
	}

	result := AllocateOverTime(events, entities)

	if !result[0].Total.Equal(decimal.Zero) {
		t.Fatalf("expected 0, got %s", result[0].Total)
	}
}

// Valida que si no existen entidades, el resultado sea un slice vacío
// y no ocurra ningún panic.
func TestNoEntities(t *testing.T) {
	t0 := time.Now()

	events := []testEvent{
		{t: t0, v: decimal.NewFromInt(100)},
	}

	result := AllocateOverTime[testEntity](events, nil)

	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d", len(result))
	}
}

// Valida que si no existen eventos, todas las entidades
// mantengan su acumulación en cero.
func TestNoEvents(t *testing.T) {
	t0 := time.Now()

	entities := []testEntity{
		{start: t0, end: t0.Add(time.Hour)},
	}

	result := AllocateOverTime[testEntity, testEvent](nil, entities)

	if !result[0].Total.Equal(decimal.Zero) {
		t.Fatalf("expected 0, got %s", result[0].Total)
	}
}

// Valida el comportamiento cuando una entidad tiene un período inválido
// (StartTime posterior a EndTime). No debe recibir ningún valor.
func TestInvalidEntityPeriod(t *testing.T) {
	t0 := time.Now()

	entities := []testEntity{
		{start: t0.Add(time.Hour), end: t0},
	}

	events := []testEvent{
		{t: t0.Add(30 * time.Minute), v: decimal.NewFromInt(100)},
	}

	result := AllocateOverTime(events, entities)

	if !result[0].Total.Equal(decimal.Zero) {
		t.Fatalf("expected 0, got %s", result[0].Total)
	}
}

// Valida que el prorrateo con redondeo a 4 decimales
// conserve exactamente la suma total original.
func TestAllocateOverTime_RoundingPreservesTotal(t *testing.T) {
	t0 := time.Now()

	entities := []testEntity{
		{start: t0, end: t0.Add(time.Hour)},
		{start: t0, end: t0.Add(time.Hour)},
		{start: t0, end: t0.Add(time.Hour)},
	}

	events := []testEvent{
		{
			t: t0.Add(10 * time.Minute),
			v: decimal.RequireFromString("100"),
		},
	}

	result := AllocateOverTime(events, entities)

	sum := decimal.Zero
	for _, r := range result {
		sum = sum.Add(r.Total)

		if r.Total.Exponent() < -4 {
			t.Fatalf("more than 4 decimals: %s", r.Total)
		}
	}

	if !sum.Equal(decimal.RequireFromString("100")) {
		t.Fatalf("expected total 100, got %s", sum)
	}
}
