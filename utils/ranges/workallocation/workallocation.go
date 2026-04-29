// Package workallocation provides a deterministic engine to calculate worked time,
// overtime classification, rest subtraction, and shift-based distribution.
//
// The package operates over time ranges (with real timestamps) and performs
// a multi-step pipeline:
//
//  1. Rest subtraction:
//     Work ranges are adjusted by removing overlapping rest periods. The result
//     is a set of effective work segments. All removed portions are tracked
//     explicitly for auditability.
//
//  2. Global accumulation and overtime classification:
//     All effective segments are ordered in a single timeline. Overtime is
//     classified globally (not per item) using configured limits:
//     - Regular: first N hours
//     - Extra25: next M hours
//     - Extra35: remaining hours
//
//     Segments may be split to fit classification boundaries.
//
//  3. Re-assignment to original items:
//     Each classified segment is mapped back to its originating work range,
//     accumulating totals per item.
//
//  4. Shift (turno) distribution:
//     Segments are intersected with shifts defined by time-of-day ranges.
//     Shifts can cross midnight and are materialized per day as needed.
//     Each item receives a per-shift breakdown.
//
// Key properties:
//
//   - Overtime is computed globally across all work ranges.
//   - Rest periods are never counted as worked time.
//   - Shift distribution is purely a spatial (time intersection) operation,
//     independent from overtime classification.
//   - Full traceability is preserved via segment-level details.
//
// Invariants:
//
//   - Horas = Regular + Extra25 + Extra35
//   - Sum of shift hours equals item hours
//   - Sum of shift extras equals item extras
//
// This package is suitable for payroll, attendance systems, and any domain
// requiring precise, auditable time allocation with overtime rules
package workallocation

import (
	"sort"
	"time"

	"github.com/sfperusacdev/identitysdk/utils/ranges"
	"github.com/user0608/goones/types"
)

type Turno interface {
	TurnoID() string
	StartTurnoTime() types.JustTime
	EndTurnoTime() types.JustTime
}

type Config struct {
	RegularLimit time.Duration
	Extra25Limit time.Duration
}

type Service[W ranges.TimeRange, R ranges.TimeRange, T Turno] struct {
	config Config
}

func NewService[W ranges.TimeRange, R ranges.TimeRange, T Turno](
	config Config,
) Service[W, R, T] {
	return Service[W, R, T]{
		config: config,
	}
}

type HourKind string

const (
	Regular HourKind = "regular"
	Extra25 HourKind = "extra25"
	Extra35 HourKind = "extra35"
)

type Allocation[W ranges.TimeRange, T Turno] struct {
	Item W

	Horas    time.Duration
	Regular  time.Duration
	Extra25  time.Duration
	Extra35  time.Duration
	Descanso time.Duration

	Turnos    []TurnoAllocation[T]
	Descansos []DescansoSegment
}

type TurnoAllocation[T Turno] struct {
	TurnoID string
	Turno   T

	Horas   time.Duration
	Extra25 time.Duration
	Extra35 time.Duration

	Segments []TurnoSegment
}

type TurnoSegment struct {
	Start time.Time
	End   time.Time
	Kind  HourKind
}

type DescansoSegment struct {
	Start time.Time
	End   time.Time
}

type effectiveSegment[W ranges.TimeRange] struct {
	ItemIndex int
	Item      W
	Start     time.Time
	End       time.Time
}

type classifiedSegment[W ranges.TimeRange] struct {
	ItemIndex int
	Item      W
	Start     time.Time
	End       time.Time
	Kind      HourKind
}

type materializedTurno[T Turno] struct {
	Turno T
	Start time.Time
	End   time.Time
}

func (s Service[W, R, T]) Calculate(
	workRanges []W,
	restRanges []R,
	turnos []T,
) []Allocation[W, T] {
	result := make([]Allocation[W, T], len(workRanges))

	for i, work := range workRanges {
		result[i] = Allocation[W, T]{
			Item: work,
		}
	}

	effectiveSegments := subtractRests(workRanges, restRanges, result)
	classifiedSegments := classifyOvertime(effectiveSegments, s.config)

	for _, segment := range classifiedSegments {
		duration := segment.End.Sub(segment.Start)

		allocation := &result[segment.ItemIndex]
		allocation.Horas += duration

		switch segment.Kind {
		case Regular:
			allocation.Regular += duration
		case Extra25:
			allocation.Extra25 += duration
		case Extra35:
			allocation.Extra35 += duration
		}

		turnoRanges := materializeTurnosForRange(turnos, segment.Start, segment.End)

		for _, turnoRange := range turnoRanges {
			overlapStart := maxTime(segment.Start, turnoRange.Start)
			overlapEnd := minTime(segment.End, turnoRange.End)

			if !overlapEnd.After(overlapStart) {
				continue
			}

			addTurnoSegment(
				allocation,
				turnoRange.Turno,
				TurnoSegment{
					Start: overlapStart,
					End:   overlapEnd,
					Kind:  segment.Kind,
				},
			)
		}
	}

	return result
}

func subtractRests[W ranges.TimeRange, R ranges.TimeRange, T Turno](
	workRanges []W,
	restRanges []R,
	result []Allocation[W, T],
) []effectiveSegment[W] {
	var effectiveSegments []effectiveSegment[W]

	for i, work := range workRanges {
		segments := []effectiveSegment[W]{
			{
				ItemIndex: i,
				Item:      work,
				Start:     work.StartTime(),
				End:       work.EndTime(),
			},
		}

		for _, rest := range restRanges {
			segments = subtractRestFromSegments(segments, rest, &result[i])
			if len(segments) == 0 {
				break
			}
		}

		effectiveSegments = append(effectiveSegments, segments...)
	}

	sort.SliceStable(effectiveSegments, func(i, j int) bool {
		return effectiveSegments[i].Start.Before(effectiveSegments[j].Start)
	})

	return effectiveSegments
}

func subtractRestFromSegments[W ranges.TimeRange, R ranges.TimeRange, T Turno](
	segments []effectiveSegment[W],
	rest R,
	allocation *Allocation[W, T],
) []effectiveSegment[W] {
	var result []effectiveSegment[W]

	restStart := rest.StartTime()
	restEnd := rest.EndTime()

	for _, segment := range segments {
		if !rangesOverlap(segment.Start, segment.End, restStart, restEnd) {
			result = append(result, segment)
			continue
		}

		descansoStart := maxTime(segment.Start, restStart)
		descansoEnd := minTime(segment.End, restEnd)

		if descansoEnd.After(descansoStart) {
			addDescansoSegment(
				allocation,
				DescansoSegment{
					Start: descansoStart,
					End:   descansoEnd,
				},
			)
		}

		if restStart.After(segment.Start) {
			result = append(result, effectiveSegment[W]{
				ItemIndex: segment.ItemIndex,
				Item:      segment.Item,
				Start:     segment.Start,
				End:       minTime(restStart, segment.End),
			})
		}

		if restEnd.Before(segment.End) {
			result = append(result, effectiveSegment[W]{
				ItemIndex: segment.ItemIndex,
				Item:      segment.Item,
				Start:     maxTime(restEnd, segment.Start),
				End:       segment.End,
			})
		}
	}

	return result
}

func addDescansoSegment[W ranges.TimeRange, T Turno](
	allocation *Allocation[W, T],
	segment DescansoSegment,
) {
	duration := segment.End.Sub(segment.Start)
	if duration <= 0 {
		return
	}

	allocation.Descanso += duration
	allocation.Descansos = append(allocation.Descansos, segment)
}

func classifyOvertime[W ranges.TimeRange](
	segments []effectiveSegment[W],
	config Config,
) []classifiedSegment[W] {
	var result []classifiedSegment[W]
	var accumulated time.Duration

	regularEnd := config.RegularLimit
	extra25End := config.RegularLimit + config.Extra25Limit

	for _, segment := range segments {
		duration := segment.End.Sub(segment.Start)

		accStart := accumulated
		accEnd := accumulated + duration

		result = append(result,
			classifySegmentPart(segment, accStart, accEnd, 0, regularEnd, Regular)...,
		)

		result = append(result,
			classifySegmentPart(segment, accStart, accEnd, regularEnd, extra25End, Extra25)...,
		)

		result = append(result,
			classifySegmentPart(segment, accStart, accEnd, extra25End, accEnd, Extra35)...,
		)

		accumulated = accEnd
	}

	return result
}

func classifySegmentPart[W ranges.TimeRange](
	segment effectiveSegment[W],
	accStart time.Duration,
	accEnd time.Duration,
	bandStart time.Duration,
	bandEnd time.Duration,
	kind HourKind,
) []classifiedSegment[W] {
	overlapStart := max(accStart, bandStart)
	overlapEnd := min(accEnd, bandEnd)

	if overlapEnd <= overlapStart {
		return nil
	}

	offsetStart := overlapStart - accStart
	offsetEnd := overlapEnd - accStart

	return []classifiedSegment[W]{
		{
			ItemIndex: segment.ItemIndex,
			Item:      segment.Item,
			Start:     segment.Start.Add(offsetStart),
			End:       segment.Start.Add(offsetEnd),
			Kind:      kind,
		},
	}
}

func materializeTurnosForRange[T Turno](
	turnos []T,
	start time.Time,
	end time.Time,
) []materializedTurno[T] {
	if !end.After(start) {
		return nil
	}

	var result []materializedTurno[T]

	startDate := dateOnly(start).AddDate(0, 0, -1)
	endDate := dateOnly(end).AddDate(0, 0, 1)

	for day := startDate; !day.After(endDate); day = day.AddDate(0, 0, 1) {
		for _, turno := range turnos {
			turnoStart := day.Add(time.Duration(turno.StartTurnoTime()))
			turnoEnd := day.Add(time.Duration(turno.EndTurnoTime()))

			if !turnoEnd.After(turnoStart) {
				turnoEnd = turnoEnd.AddDate(0, 0, 1)
			}

			if rangesOverlap(start, end, turnoStart, turnoEnd) {
				result = append(result, materializedTurno[T]{
					Turno: turno,
					Start: turnoStart,
					End:   turnoEnd,
				})
			}
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Start.Before(result[j].Start)
	})

	return result
}

func addTurnoSegment[W ranges.TimeRange, T Turno](
	allocation *Allocation[W, T],
	turno T,
	segment TurnoSegment,
) {
	duration := segment.End.Sub(segment.Start)
	if duration <= 0 {
		return
	}

	for i := range allocation.Turnos {
		if allocation.Turnos[i].TurnoID == turno.TurnoID() {
			addDurationByKind(&allocation.Turnos[i], segment.Kind, duration)
			allocation.Turnos[i].Segments = append(allocation.Turnos[i].Segments, segment)
			return
		}
	}

	turnoAllocation := TurnoAllocation[T]{
		TurnoID:  turno.TurnoID(),
		Turno:    turno,
		Segments: []TurnoSegment{segment},
	}

	addDurationByKind(&turnoAllocation, segment.Kind, duration)

	allocation.Turnos = append(allocation.Turnos, turnoAllocation)
}

func addDurationByKind[T Turno](
	allocation *TurnoAllocation[T],
	kind HourKind,
	duration time.Duration,
) {
	allocation.Horas += duration

	switch kind {
	case Extra25:
		allocation.Extra25 += duration
	case Extra35:
		allocation.Extra35 += duration
	}
}

func rangesOverlap(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}

	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}

	return b
}
