// Package rangeoverlap provides utilities to calculate the time overlap
// between two sets of time ranges.
//
// It is intended for scenarios such as computing how much time a person
// spent in a secondary state (e.g. rest, break, pause) during a primary
// period (e.g. work, shift, session).
//
// Given two collections of time ranges:
//
//   - Primary ranges (e.g. work schedules)
//   - Secondary ranges (e.g. rest periods)
//
// The package computes, for each primary range, the total duration of
// overlap with all secondary ranges.
//
// The overlap between two ranges is defined as:
//
//	overlap = min(endA, endB) - max(startA, startB)
//
// If the result is negative or zero, no overlap exists.
//
// The result is expressed as a duration per primary range, without
// modifying the original ranges.
//
// Preconditions:
//   - Input ranges must be valid (StartTime <= EndTime)
//   - Ranges within each collection must not overlap with each other
//   - Ranges are expected to be logically consistent (already validated)
//
// This design keeps the computation simple, deterministic, and focused
// solely on overlap calculation, leaving validation and normalization
// to upstream logic.
package rangeoverlap

import (
	"testing"
	"time"
)

type testRange struct {
	name  string
	start time.Time
	end   time.Time
}

func (r testRange) StartTime() time.Time { return r.start }
func (r testRange) EndTime() time.Time   { return r.end }

func at(hour, minute int) time.Time {
	return time.Date(2026, 1, 1, hour, minute, 0, 0, time.UTC)
}

func TestCalculateOverlapDurations(t *testing.T) {
	tests := []struct {
		name     string
		work     []testRange
		rest     []testRange
		expected []time.Duration
	}{
		{
			name: "single work range with one rest fully inside",
			work: []testRange{
				{name: "work-1", start: at(8, 0), end: at(12, 0)},
			},
			rest: []testRange{
				{name: "rest-1", start: at(9, 0), end: at(9, 30)},
			},
			expected: []time.Duration{
				30 * time.Minute,
			},
		},
		{
			name: "single work range with multiple rests inside",
			work: []testRange{
				{name: "work-1", start: at(8, 0), end: at(12, 0)},
			},
			rest: []testRange{
				{name: "rest-1", start: at(9, 0), end: at(9, 30)},
				{name: "rest-2", start: at(10, 0), end: at(10, 15)},
				{name: "rest-3", start: at(11, 0), end: at(11, 45)},
			},
			expected: []time.Duration{
				90 * time.Minute,
			},
		},
		{
			name: "rest overlaps left edge of work",
			work: []testRange{
				{name: "work-1", start: at(8, 0), end: at(12, 0)},
			},
			rest: []testRange{
				{name: "rest-1", start: at(7, 30), end: at(8, 30)},
			},
			expected: []time.Duration{
				30 * time.Minute,
			},
		},
		{
			name: "rest overlaps right edge of work",
			work: []testRange{
				{name: "work-1", start: at(8, 0), end: at(12, 0)},
			},
			rest: []testRange{
				{name: "rest-1", start: at(11, 30), end: at(12, 30)},
			},
			expected: []time.Duration{
				30 * time.Minute,
			},
		},
		{
			name: "rest fully covers work",
			work: []testRange{
				{name: "work-1", start: at(8, 0), end: at(12, 0)},
			},
			rest: []testRange{
				{name: "rest-1", start: at(7, 0), end: at(13, 0)},
			},
			expected: []time.Duration{
				4 * time.Hour,
			},
		},
		{
			name: "rest touches work start but does not overlap",
			work: []testRange{
				{name: "work-1", start: at(8, 0), end: at(12, 0)},
			},
			rest: []testRange{
				{name: "rest-1", start: at(7, 0), end: at(8, 0)},
			},
			expected: []time.Duration{
				0,
			},
		},
		{
			name: "rest touches work end but does not overlap",
			work: []testRange{
				{name: "work-1", start: at(8, 0), end: at(12, 0)},
			},
			rest: []testRange{
				{name: "rest-1", start: at(12, 0), end: at(13, 0)},
			},
			expected: []time.Duration{
				0,
			},
		},
		{
			name: "rest before work does not overlap",
			work: []testRange{
				{name: "work-1", start: at(8, 0), end: at(12, 0)},
			},
			rest: []testRange{
				{name: "rest-1", start: at(6, 0), end: at(7, 0)},
			},
			expected: []time.Duration{
				0,
			},
		},
		{
			name: "rest after work does not overlap",
			work: []testRange{
				{name: "work-1", start: at(8, 0), end: at(12, 0)},
			},
			rest: []testRange{
				{name: "rest-1", start: at(13, 0), end: at(14, 0)},
			},
			expected: []time.Duration{
				0,
			},
		},
		{
			name: "multiple work ranges with shared rest ranges",
			work: []testRange{
				{name: "work-1", start: at(8, 0), end: at(12, 0)},
				{name: "work-2", start: at(13, 0), end: at(17, 0)},
			},
			rest: []testRange{
				{name: "rest-1", start: at(9, 0), end: at(9, 30)},
				{name: "rest-2", start: at(11, 30), end: at(13, 30)},
				{name: "rest-3", start: at(15, 0), end: at(15, 45)},
			},
			expected: []time.Duration{
				60 * time.Minute, // 30m + 30m
				75 * time.Minute, // 30m + 45m
			},
		},
		{
			name: "empty work ranges",
			work: nil,
			rest: []testRange{
				{name: "rest-1", start: at(9, 0), end: at(10, 0)},
			},
			expected: []time.Duration{},
		},
		{
			name: "empty rest ranges",
			work: []testRange{
				{name: "work-1", start: at(8, 0), end: at(12, 0)},
			},
			rest: nil,
			expected: []time.Duration{
				0,
			},
		},
		{
			name: "zero duration rest does not overlap",
			work: []testRange{
				{name: "work-1", start: at(8, 0), end: at(12, 0)},
			},
			rest: []testRange{
				{name: "rest-1", start: at(9, 0), end: at(9, 0)},
			},
			expected: []time.Duration{
				0,
			},
		},
		{
			name: "zero duration work does not accumulate overlap",
			work: []testRange{
				{name: "work-1", start: at(8, 0), end: at(8, 0)},
			},
			rest: []testRange{
				{name: "rest-1", start: at(7, 0), end: at(9, 0)},
			},
			expected: []time.Duration{
				0,
			},
		},
		{
			name: "work crossing midnight",
			work: []testRange{
				{
					name:  "work-1",
					start: time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC),
					end:   time.Date(2026, 1, 2, 6, 0, 0, 0, time.UTC),
				},
			},
			rest: []testRange{
				{
					name:  "rest-1",
					start: time.Date(2026, 1, 2, 1, 0, 0, 0, time.UTC),
					end:   time.Date(2026, 1, 2, 1, 30, 0, 0, time.UTC),
				},
			},
			expected: []time.Duration{
				30 * time.Minute,
			},
		},
		{
			name: "overlapping rest ranges are counted independently",
			work: []testRange{
				{name: "work-1", start: at(8, 0), end: at(12, 0)},
			},
			rest: []testRange{
				{name: "rest-1", start: at(9, 0), end: at(10, 0)},
				{name: "rest-2", start: at(9, 30), end: at(10, 30)},
			},
			expected: []time.Duration{
				2 * time.Hour,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateOverlapDurations(tt.work, tt.rest)

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d results, got %d", len(tt.expected), len(result))
			}

			for i := range result {
				if result[i].OverlapDuration != tt.expected[i] {
					t.Fatalf(
						"item %d expected overlap duration %s, got %s",
						i,
						tt.expected[i],
						result[i].OverlapDuration,
					)
				}
			}
		})
	}
}
