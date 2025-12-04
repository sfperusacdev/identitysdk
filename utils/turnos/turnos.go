package turnos

import (
	"log/slog"
	"sort"
	"time"
)

type Turno struct {
	Codigo           string `json:"codigo"`
	Descripcion      string `json:"descripcion"`
	Inicio           string `json:"inicio"`
	Fin              string `json:"fin"`
	RefereciaExterna string `json:"referecia_externa"`
}

type turnoParsed struct {
	Codigo    string
	Start     time.Duration
	End       time.Duration
	_Original Turno
}

type Segmentable interface {
	GetStartTime() time.Time
	GetEndTime() time.Time
	NewSegment(turno Turno, start, end time.Time) Segmentable
}

func preprocesarTurnos(turnos []Turno) []turnoParsed {
	parsed := make([]turnoParsed, 0, len(turnos))

	for _, t := range turnos {
		hi, err1 := time.Parse(time.TimeOnly, t.Inicio)
		hf, err2 := time.Parse(time.TimeOnly, t.Fin)
		if err1 != nil || err2 != nil {
			continue
		}

		startDur := time.Duration(hi.Hour())*time.Hour +
			time.Duration(hi.Minute())*time.Minute +
			time.Duration(hi.Second())*time.Second

		endDur := time.Duration(hf.Hour())*time.Hour +
			time.Duration(hf.Minute())*time.Minute +
			time.Duration(hf.Second())*time.Second

		parsed = append(parsed, turnoParsed{
			Codigo:    t.Codigo,
			Start:     startDur,
			End:       endDur,
			_Original: t,
		})
	}

	sort.Slice(parsed, func(i, j int) bool { return parsed[i].Start < parsed[j].Start })
	return parsed
}

func truncMidnight(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
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

func segmentarPorTurnosParsed[T Segmentable](turnos []turnoParsed, origen T, holgura time.Duration) []T {
	globalStart := origen.GetStartTime()
	globalEnd := origen.GetEndTime()

	if !globalStart.Before(globalEnd) {
		slog.Warn("invalid time range: start >= end")
		return nil
	}

	startDay := truncMidnight(globalStart).AddDate(0, 0, -1)
	endDay := truncMidnight(globalEnd)

	type wrap struct {
		seg       T
		codigo    string
		_Original Turno
	}

	var wraps []wrap

	for day := startDay; !day.After(endDay); day = day.AddDate(0, 0, 1) {
		for _, pt := range turnos {

			shiftStart := day.Add(pt.Start)
			shiftEnd := day.Add(pt.End)

			if !shiftEnd.After(shiftStart) {
				shiftEnd = shiftEnd.Add(24 * time.Hour)
			}

			segStart := maxTime(globalStart, shiftStart)
			segEnd := minTime(globalEnd, shiftEnd)

			if segStart.Before(segEnd) {
				segI := origen.NewSegment(pt._Original, segStart, segEnd)
				if segI == nil {
					continue
				}
				seg, ok := segI.(T)
				if !ok {
					continue
				}
				wraps = append(wraps, wrap{seg: seg, codigo: pt.Codigo, _Original: pt._Original})
			}
		}
	}

	if len(wraps) == 0 {
		slog.Warn("no segments generated")
		return nil
	}

	if holgura > 0 && len(wraps) >= 2 {
		lastIdx := len(wraps) - 1
		prev := wraps[lastIdx-1]
		last := wraps[lastIdx]

		tailDur := last.seg.GetEndTime().Sub(last.seg.GetStartTime())

		if tailDur > 0 && tailDur <= holgura {
			mergedI := origen.NewSegment(prev._Original, prev.seg.GetStartTime(), last.seg.GetEndTime())
			if mergedI != nil {
				if merged, ok := mergedI.(T); ok {
					wraps[lastIdx-1] = wrap{seg: merged, codigo: prev.codigo}
					wraps = wraps[:lastIdx]
				}
			}
		}
	}

	result := make([]T, len(wraps))
	for i := range wraps {
		result[i] = wraps[i].seg
	}

	return result
}

func SegmentarPorTurnos[T Segmentable](turnos []Turno, origen T, holguraMin int) []T {
	parsed := preprocesarTurnos(turnos)
	if len(parsed) == 0 {
		slog.Warn("empty or invalid turno list")
		return nil
	}

	return segmentarPorTurnosParsed(parsed, origen, time.Duration(holguraMin)*time.Minute)
}
