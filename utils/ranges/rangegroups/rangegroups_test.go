package rangegroups

import (
	"reflect"
	"testing"
	"time"
)

type testRange struct {
	id    string
	start time.Time
	end   time.Time
}

func (r testRange) ID() string           { return r.id }
func (r testRange) StartTime() time.Time { return r.start }
func (r testRange) EndTime() time.Time   { return r.end }

func at(hour, minute int) time.Time {
	return time.Date(2026, 1, 1, hour, minute, 0, 0, time.UTC)
}

func atDay(day, hour, minute int) time.Time {
	return time.Date(2026, 1, day, hour, minute, 0, 0, time.UTC)
}

func TestFindGroup(t *testing.T) {
	tests := []struct {
		name   string
		items  []testRange
		id     string
		maxGap time.Duration
		want   []string
		wantOK bool
	}{
		{
			name: "connects both directions",
			items: []testRange{
				{id: "a", start: at(8, 0), end: at(9, 0)},
				{id: "b", start: at(9, 10), end: at(10, 0)},
				{id: "c", start: at(10, 20), end: at(11, 0)},
				{id: "d", start: at(12, 0), end: at(13, 0)},
			},
			id:     "b",
			maxGap: 20 * time.Minute,
			want:   []string{"a", "b", "c"},
			wantOK: true,
		},
		{
			name: "stops when gap is greater than max",
			items: []testRange{
				{id: "a", start: at(8, 0), end: at(9, 0)},
				{id: "b", start: at(9, 31), end: at(10, 0)},
				{id: "c", start: at(10, 15), end: at(11, 0)},
			},
			id:     "b",
			maxGap: 30 * time.Minute,
			want:   []string{"b", "c"},
			wantOK: true,
		},
		{
			name: "includes gap equal to max",
			items: []testRange{
				{id: "a", start: at(8, 0), end: at(9, 0)},
				{id: "b", start: at(9, 30), end: at(10, 0)},
			},
			id:     "a",
			maxGap: 30 * time.Minute,
			want:   []string{"a", "b"},
			wantOK: true,
		},
		{
			name: "overlapping ranges are connected",
			items: []testRange{
				{id: "a", start: at(8, 0), end: at(10, 0)},
				{id: "b", start: at(9, 0), end: at(11, 0)},
			},
			id:     "b",
			maxGap: 0,
			want:   []string{"a", "b"},
			wantOK: true,
		},
		{
			name: "overlaps with noise on both sides",
			items: []testRange{
				{id: "noise-left", start: at(6, 0), end: at(7, 0)},
				{id: "a", start: at(8, 0), end: at(9, 30)},
				{id: "b", start: at(9, 0), end: at(10, 0)},
				{id: "c", start: at(9, 45), end: at(11, 0)},
				{id: "d", start: at(11, 10), end: at(12, 0)},
				{id: "noise-right", start: at(14, 0), end: at(15, 0)},
			},
			id:     "b",
			maxGap: 15 * time.Minute,
			want:   []string{"a", "b", "c", "d"},
			wantOK: true,
		},
		{
			name: "separate overlapping clusters only returns selected cluster",
			items: []testRange{
				{id: "noise-a", start: at(6, 0), end: at(8, 0)},
				{id: "noise-b", start: at(7, 30), end: at(8, 30)},
				{id: "a", start: at(10, 0), end: at(11, 30)},
				{id: "b", start: at(11, 0), end: at(12, 0)},
				{id: "c", start: at(12, 5), end: at(13, 0)},
				{id: "noise-c", start: at(15, 0), end: at(17, 0)},
				{id: "noise-d", start: at(16, 30), end: at(18, 0)},
			},
			id:     "b",
			maxGap: 10 * time.Minute,
			want:   []string{"a", "b", "c"},
			wantOK: true,
		},
		{
			name: "mixed overlaps and allowed gaps across noisy timeline",
			items: []testRange{
				{id: "noise-left", start: at(5, 0), end: at(6, 0)},
				{id: "a", start: at(8, 0), end: at(9, 0)},
				{id: "b", start: at(8, 45), end: at(10, 0)},
				{id: "c", start: at(10, 20), end: at(11, 0)},
				{id: "d", start: at(10, 50), end: at(12, 0)},
				{id: "e", start: at(12, 20), end: at(13, 0)},
				{id: "noise-right", start: at(13, 45), end: at(14, 0)},
			},
			id:     "d",
			maxGap: 20 * time.Minute,
			want:   []string{"a", "b", "c", "d", "e"},
			wantOK: true,
		},
		{
			name: "returns false when group exceeds 24 hours",
			items: []testRange{
				{id: "a", start: atDay(1, 8, 0), end: atDay(1, 12, 0)},
				{id: "b", start: atDay(1, 12, 10), end: atDay(1, 18, 0)},
				{id: "c", start: atDay(1, 18, 10), end: atDay(2, 8, 1)},
			},
			id:     "b",
			maxGap: 10 * time.Minute,
			want:   nil,
			wantOK: false,
		},
		{
			name: "allows group with exactly 24 hours",
			items: []testRange{
				{id: "a", start: atDay(1, 8, 0), end: atDay(1, 12, 0)},
				{id: "b", start: atDay(1, 12, 0), end: atDay(1, 18, 0)},
				{id: "c", start: atDay(1, 18, 0), end: atDay(2, 8, 0)},
			},
			id:     "b",
			maxGap: 0,
			want:   []string{"a", "b", "c"},
			wantOK: true,
		},
		{
			name: "id not found returns nil",
			items: []testRange{
				{id: "a", start: at(8, 0), end: at(9, 0)},
			},
			id:     "missing",
			maxGap: time.Hour,
			want:   nil,
			wantOK: false,
		},
		{
			name:   "empty list returns nil",
			items:  nil,
			id:     "a",
			maxGap: time.Hour,
			want:   nil,
			wantOK: false,
		},
		{
			name: "negative max gap returns initial range only for non-overlapping ranges",
			items: []testRange{
				{id: "a", start: at(8, 0), end: at(9, 0)},
				{id: "b", start: at(9, 0), end: at(10, 0)},
			},
			id:     "a",
			maxGap: -time.Minute,
			want:   []string{"a"},
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, ok := FindGroup(tt.items, tt.id, tt.maxGap)
			got := rangeIDs(group)

			if ok != tt.wantOK {
				t.Fatalf("expected ok %v, got %v", tt.wantOK, ok)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func rangeIDs(items []testRange) []string {
	if items == nil {
		return nil
	}

	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID())
	}

	return ids
}
