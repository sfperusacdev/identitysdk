// Package workallocation provides utilities to detect overlaps between time ranges.
//
// It operates on types implementing the TimeRange interface and allows marking
// which ranges are in conflict with at least one other range in a given collection.
//
// Overlap detection uses strict interval comparison:
// a range A overlaps B if A.start < B.end && B.start < A.end.
//
// Adjacent ranges where end == start are not considered overlapping.
package workallocation

import (
	"testing"
	"time"
)

type mockRange struct {
	start time.Time
	end   time.Time
}

func (m mockRange) StartTime() time.Time { return m.start }
func (m mockRange) EndTime() time.Time   { return m.end }

func TestMarkOverlappingRanges(t *testing.T) {
	cases := []struct {
		name     string
		ranges   []mockRange
		expected []bool
	}{
		{
			name: "no overlap",
			ranges: []mockRange{
				{time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 11, 0, 0, 0, time.UTC)},
				{time.Date(2026, 4, 24, 11, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)},
				{time.Date(2026, 4, 24, 13, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 14, 0, 0, 0, time.UTC)},
			},
			expected: []bool{false, false, false},
		},
		{
			name: "simple overlap",
			ranges: []mockRange{
				{time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 11, 0, 0, 0, time.UTC)},
				{time.Date(2026, 4, 24, 10, 30, 0, 0, time.UTC), time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)},
			},
			expected: []bool{true, true},
		},
		{
			name: "chain overlap",
			ranges: []mockRange{
				{time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)},
				{time.Date(2026, 4, 24, 11, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 13, 0, 0, 0, time.UTC)},
				{time.Date(2026, 4, 24, 12, 30, 0, 0, time.UTC), time.Date(2026, 4, 24, 14, 0, 0, 0, time.UTC)},
			},
			expected: []bool{true, true, true},
		},
		{
			name: "contained",
			ranges: []mockRange{
				{time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 15, 0, 0, 0, time.UTC)},
				{time.Date(2026, 4, 24, 11, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)},
				{time.Date(2026, 4, 24, 16, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 17, 0, 0, 0, time.UTC)},
			},
			expected: []bool{true, true, false},
		},
		{
			name: "same start",
			ranges: []mockRange{
				{time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 11, 0, 0, 0, time.UTC)},
				{time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)},
			},
			expected: []bool{true, true},
		},
		{
			name: "same end",
			ranges: []mockRange{
				{time.Date(2026, 4, 24, 9, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 11, 0, 0, 0, time.UTC)},
				{time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 11, 0, 0, 0, time.UTC)},
			},
			expected: []bool{true, true},
		},
		{
			name: "different days",
			ranges: []mockRange{
				{time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 11, 0, 0, 0, time.UTC)},
				{time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC), time.Date(2026, 4, 25, 11, 0, 0, 0, time.UTC)},
			},
			expected: []bool{false, false},
		},
		{
			name: "zero length",
			ranges: []mockRange{
				{time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)},
				{time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 11, 0, 0, 0, time.UTC)},
			},
			expected: []bool{false, false},
		},
		{
			name: "multiple groups",
			ranges: []mockRange{
				{time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 11, 0, 0, 0, time.UTC)},
				{time.Date(2026, 4, 24, 10, 30, 0, 0, time.UTC), time.Date(2026, 4, 24, 11, 30, 0, 0, time.UTC)},
				{time.Date(2026, 4, 24, 13, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 14, 0, 0, 0, time.UTC)},
				{time.Date(2026, 4, 24, 13, 30, 0, 0, time.UTC), time.Date(2026, 4, 24, 14, 30, 0, 0, time.UTC)},
				{time.Date(2026, 4, 24, 16, 0, 0, 0, time.UTC), time.Date(2026, 4, 24, 17, 0, 0, 0, time.UTC)},
			},
			expected: []bool{true, true, true, true, false},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := MarkOverlappingRanges(c.ranges)

			if len(res) != len(c.expected) {
				t.Fatalf("length mismatch: got %d want %d", len(res), len(c.expected))
			}

			for i := range res {
				if res[i].Overlapping != c.expected[i] {
					t.Errorf("index %d: got %v want %v", i, res[i].Overlapping, c.expected[i])
				}
			}
		})
	}
}
