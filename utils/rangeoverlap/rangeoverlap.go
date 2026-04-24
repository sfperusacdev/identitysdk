package rangeoverlap

import "time"

type TimeRange interface {
	StartTime() time.Time
	EndTime() time.Time
}

type OverlapResult[T TimeRange] struct {
	Item            T
	OverlapDuration time.Duration
}

func CalculateOverlapDurations[T TimeRange, K TimeRange](
	primary []T,
	secondary []K,
) []OverlapResult[T] {
	results := make([]OverlapResult[T], len(primary))

	for i, item := range primary {
		var total time.Duration

		for _, other := range secondary {
			total += overlapDuration(
				item.StartTime(),
				item.EndTime(),
				other.StartTime(),
				other.EndTime(),
			)
		}

		results[i] = OverlapResult[T]{
			Item:            item,
			OverlapDuration: total,
		}
	}

	return results
}

func overlapDuration(
	aStart time.Time,
	aEnd time.Time,
	bStart time.Time,
	bEnd time.Time,
) time.Duration {
	start := maxTime(aStart, bStart)
	end := minTime(aEnd, bEnd)

	if !end.After(start) {
		return 0
	}

	return end.Sub(start)
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
