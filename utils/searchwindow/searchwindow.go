// Package searchwindow provides utilities to compute flexible search tolerances
// around time ranges.
//
// It is designed for scenarios like work schedules where events (e.g. employee
// check-ins) may occur slightly before or after the defined time range.
//
// Given an ordered, non-overlapping list of time ranges, the package calculates
// a tolerance window for each range based on the gap between adjacent ranges.
// Each gap is proportionally distributed between neighboring ranges using
// configurable percentages.
//
// For each pair of consecutive ranges:
//
//	gap = next.StartTime() - current.EndTime()
//
// The gap is split into:
//   - AfterEnd (expansion to the right of the current range)
//   - BeforeStart (expansion to the left of the next range)
//
// Each side is limited by a configurable maximum duration.
//
// The result does not directly modify time ranges. Instead, it returns
// durations (tolerances) that can later be applied:
//
//	searchStart = StartTime() - BeforeStart
//	searchEnd   = EndTime() + AfterEnd
//
// Preconditions:
//   - Input ranges must be sorted by StartTime()
//   - Ranges must not overlap
//
// This design keeps the calculation generic, reusable, and independent of
// how the final search window is applied.
package searchwindow

import (
	"errors"
	"time"
)

type TimeRange interface {
	StartTime() time.Time
	EndTime() time.Time
}

type ToleranceConfig struct {
	BeforePercent float64
	AfterPercent  float64
	MaxBefore     time.Duration
	MaxAfter      time.Duration
}

type Tolerance struct {
	BeforeStart time.Duration
	AfterEnd    time.Duration
}

type SearchTolerance[T TimeRange] struct {
	Item      T
	Tolerance Tolerance
}

var (
	ErrNegativePercent     = errors.New("percent values cannot be negative")
	ErrInvalidPercentTotal = errors.New("before percent plus after percent cannot be greater than 1")
	ErrNegativeMax         = errors.New("max tolerance values cannot be negative")
)

func (c ToleranceConfig) Validate() error {
	if c.BeforePercent < 0 || c.AfterPercent < 0 {
		return ErrNegativePercent
	}

	if c.BeforePercent+c.AfterPercent > 1 {
		return ErrInvalidPercentTotal
	}

	if c.MaxBefore < 0 || c.MaxAfter < 0 {
		return ErrNegativeMax
	}

	return nil
}

func CalculateSearchTolerances[T TimeRange](
	items []T,
	config ToleranceConfig,
) ([]SearchTolerance[T], error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	results := make([]SearchTolerance[T], len(items))

	for i, item := range items {
		results[i] = SearchTolerance[T]{
			Item:      item,
			Tolerance: Tolerance{},
		}
	}

	for i := 0; i < len(items)-1; i++ {
		current := items[i]
		next := items[i+1]

		gap := next.StartTime().Sub(current.EndTime())
		if gap <= 0 {
			continue
		}

		results[i].Tolerance.AfterEnd = calculateTolerance(
			gap,
			config.AfterPercent,
			config.MaxAfter,
		)

		results[i+1].Tolerance.BeforeStart = calculateTolerance(
			gap,
			config.BeforePercent,
			config.MaxBefore,
		)
	}

	return results, nil
}

func calculateTolerance(
	gap time.Duration,
	percent float64,
	max time.Duration,
) time.Duration {
	if gap <= 0 || percent <= 0 || max == 0 {
		return 0
	}

	value := time.Duration(float64(gap) * percent)

	if value > max {
		return max
	}

	return value
}
