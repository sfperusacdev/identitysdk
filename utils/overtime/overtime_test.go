package overtime

import (
	"testing"
	"time"
)

type testItem struct {
	name  string
	hours time.Duration
}

func (i testItem) Hours() time.Duration {
	return i.hours
}

func TestCalculate(t *testing.T) {
	cfg := Config{
		RegularLimit: time.Hour * 8,
		Extra20Limit: time.Hour * 2,
	}

	tests := []struct {
		name     string
		items    []testItem
		expected []Allocation[testItem]
	}{
		{
			name: "total less than regular limit",
			items: []testItem{
				{name: "A", hours: 3 * time.Hour},
				{name: "B", hours: 4 * time.Hour},
			},
			expected: []Allocation[testItem]{
				{Item: testItem{name: "A", hours: 3 * time.Hour}, Regular: 3 * time.Hour},
				{Item: testItem{name: "B", hours: 4 * time.Hour}, Regular: 4 * time.Hour},
			},
		},
		{
			name: "total exactly regular limit",
			items: []testItem{
				{name: "A", hours: 3 * time.Hour},
				{name: "B", hours: 5 * time.Hour},
			},
			expected: []Allocation[testItem]{
				{Item: testItem{name: "A", hours: 3 * time.Hour}, Regular: 3 * time.Hour},
				{Item: testItem{name: "B", hours: 5 * time.Hour}, Regular: 5 * time.Hour},
			},
		},
		{
			name: "item crosses from regular to extra20",
			items: []testItem{
				{name: "A", hours: 7 * time.Hour},
				{name: "B", hours: 2 * time.Hour},
			},
			expected: []Allocation[testItem]{
				{Item: testItem{name: "A", hours: 7 * time.Hour}, Regular: 7 * time.Hour},
				{Item: testItem{name: "B", hours: 2 * time.Hour}, Regular: 1 * time.Hour, Extra20: 1 * time.Hour},
			},
		},
		{
			name: "item crosses from extra20 to extra30",
			items: []testItem{
				{name: "A", hours: 8 * time.Hour},
				{name: "B", hours: 3 * time.Hour},
			},
			expected: []Allocation[testItem]{
				{Item: testItem{name: "A", hours: 8 * time.Hour}, Regular: 8 * time.Hour},
				{Item: testItem{name: "B", hours: 3 * time.Hour}, Extra20: 2 * time.Hour, Extra30: 1 * time.Hour},
			},
		},
		{
			name: "single item crosses all bands",
			items: []testItem{
				{name: "A", hours: 12 * time.Hour},
			},
			expected: []Allocation[testItem]{
				{Item: testItem{name: "A", hours: 12 * time.Hour}, Regular: 8 * time.Hour, Extra20: 2 * time.Hour, Extra30: 2 * time.Hour},
			},
		},
		{
			name: "example A3 B4 C3 D4",
			items: []testItem{
				{name: "A", hours: 3 * time.Hour},
				{name: "B", hours: 4 * time.Hour},
				{name: "C", hours: 3 * time.Hour},
				{name: "D", hours: 4 * time.Hour},
			},
			expected: []Allocation[testItem]{
				{Item: testItem{name: "A", hours: 3 * time.Hour}, Regular: 3 * time.Hour},
				{Item: testItem{name: "B", hours: 4 * time.Hour}, Regular: 4 * time.Hour},
				{Item: testItem{name: "C", hours: 3 * time.Hour}, Regular: 1 * time.Hour, Extra20: 2 * time.Hour},
				{Item: testItem{name: "D", hours: 4 * time.Hour}, Extra30: 4 * time.Hour},
			},
		},
		{
			name: "multiple items fully inside extra30",
			items: []testItem{
				{name: "A", hours: 8 * time.Hour},
				{name: "B", hours: 2 * time.Hour},
				{name: "C", hours: 1 * time.Hour},
				{name: "D", hours: 3 * time.Hour},
			},
			expected: []Allocation[testItem]{
				{Item: testItem{name: "A", hours: 8 * time.Hour}, Regular: 8 * time.Hour},
				{Item: testItem{name: "B", hours: 2 * time.Hour}, Extra20: 2 * time.Hour},
				{Item: testItem{name: "C", hours: 1 * time.Hour}, Extra30: 1 * time.Hour},
				{Item: testItem{name: "D", hours: 3 * time.Hour}, Extra30: 3 * time.Hour},
			},
		},
		{
			name:     "empty input",
			items:    []testItem{},
			expected: []Allocation[testItem]{},
		},
		{
			name: "minutes precision",
			items: []testItem{
				{name: "A", hours: 7*time.Hour + 30*time.Minute},
				{name: "B", hours: 45 * time.Minute},
				{name: "C", hours: 2 * time.Hour},
			},
			expected: []Allocation[testItem]{
				{Item: testItem{name: "A", hours: 7*time.Hour + 30*time.Minute}, Regular: 7*time.Hour + 30*time.Minute},
				{Item: testItem{name: "B", hours: 45 * time.Minute}, Regular: 30 * time.Minute, Extra20: 15 * time.Minute},
				{Item: testItem{name: "C", hours: 2 * time.Hour}, Extra20: 105 * time.Minute, Extra30: 15 * time.Minute},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Calculate(tt.items, cfg)

			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d results, got %d", len(tt.expected), len(got))
			}

			for i := range got {
				if got[i].Regular != tt.expected[i].Regular {
					t.Fatalf("item %d expected regular %s, got %s", i, tt.expected[i].Regular, got[i].Regular)
				}

				if got[i].Extra20 != tt.expected[i].Extra20 {
					t.Fatalf("item %d expected extra20 %s, got %s", i, tt.expected[i].Extra20, got[i].Extra20)
				}

				if got[i].Extra30 != tt.expected[i].Extra30 {
					t.Fatalf("item %d expected extra30 %s, got %s", i, tt.expected[i].Extra30, got[i].Extra30)
				}
			}
		})
	}
}
