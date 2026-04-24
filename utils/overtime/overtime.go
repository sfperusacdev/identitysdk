// Package overtime provides utilities to classify worked time into
// regular hours and overtime based on accumulated duration.
//
// It is designed for labor scenarios where compensation depends on
// thresholds defined by regulation. For example:
//
//   - Up to a base limit (e.g. 8 hours): regular hours
//   - Next segment (e.g. 2 hours): overtime at +20%
//   - Remaining time: overtime at +30%
//
// The package processes a sequence of items that expose a duration
// via the HourItem interface. It accumulates time from the beginning
// and splits each item across the defined thresholds.
//
// Conceptually, the accumulated timeline is divided into segments:
//
//	0 ───────── RegularLimit ───────── RegularLimit+Extra20Limit ───────── ∞
//	   Regular              Extra20                          Extra30
//
// Each item contributes to one or more segments depending on its
// position in the accumulated timeline.
//
// Preconditions:
//   - Items are assumed to be in chronological/logical order
//   - Each item must return a non-negative duration
//
// The result preserves the original items and assigns, for each one,
// how much time belongs to:
//   - Regular hours
//   - First overtime segment (e.g. +20%)
//   - Remaining overtime (e.g. +30%)
//
// This approach is deterministic, extensible, and aligned with
// regulation-based time classification.
package overtime

import "time"

type HourItem interface {
	Hours() time.Duration
}

type Config struct {
	RegularLimit time.Duration
	Extra20Limit time.Duration
}

type Allocation[T HourItem] struct {
	Item    T
	Regular time.Duration
	Extra20 time.Duration
	Extra30 time.Duration
}

func Calculate[T HourItem](items []T, cfg Config) []Allocation[T] {
	results := make([]Allocation[T], len(items))

	var acc time.Duration

	for i, item := range items {
		h := item.Hours()

		start := acc
		end := acc + h

		results[i] = Allocation[T]{
			Item:    item,
			Regular: overlap(start, end, 0, cfg.RegularLimit),
			Extra20: overlap(start, end,
				cfg.RegularLimit,
				cfg.RegularLimit+cfg.Extra20Limit,
			),
			Extra30: overlap(start, end,
				cfg.RegularLimit+cfg.Extra20Limit,
				end, // infinito práctico
			),
		}

		acc = end
	}

	return results
}

func overlap(aStart, aEnd, bStart, bEnd time.Duration) time.Duration {
	start := max(aStart, bStart)
	end := min(aEnd, bEnd)

	if end <= start {
		return 0
	}

	return end - start
}
