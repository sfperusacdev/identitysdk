package searchwindow

import (
	"errors"
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

func TestCalculateSearchTolerances_WithExpectedSearchWindow(t *testing.T) {
	config := ToleranceConfig{
		BeforePercent: 0.30,
		AfterPercent:  0.70,
		MaxBefore:     time.Hour,
		MaxAfter:      time.Hour,
	}

	tests := []struct {
		name     string
		items    []testRange
		expected []struct {
			before     time.Duration
			after      time.Duration
			searchFrom time.Time
			searchTo   time.Time
		}
	}{
		{
			name: "two ranges with 2h gap and max cap applied",
			items: []testRange{
				{name: "r1", start: at(9, 0), end: at(10, 0)},
				{name: "r2", start: at(12, 0), end: at(13, 0)},
			},
			expected: []struct {
				before     time.Duration
				after      time.Duration
				searchFrom time.Time
				searchTo   time.Time
			}{
				{
					before:     0,
					after:      time.Hour,
					searchFrom: at(9, 0),
					searchTo:   at(11, 0),
				},
				{
					before:     36 * time.Minute,
					after:      0,
					searchFrom: at(11, 24),
					searchTo:   at(13, 0),
				},
			},
		},
		{
			name: "two ranges with 1h gap without cap",
			items: []testRange{
				{name: "r1", start: at(8, 0), end: at(9, 0)},
				{name: "r2", start: at(10, 0), end: at(11, 0)},
			},
			expected: []struct {
				before     time.Duration
				after      time.Duration
				searchFrom time.Time
				searchTo   time.Time
			}{
				{
					before:     0,
					after:      42 * time.Minute,
					searchFrom: at(8, 0),
					searchTo:   at(9, 42),
				},
				{
					before:     18 * time.Minute,
					after:      0,
					searchFrom: at(9, 42),
					searchTo:   at(11, 0),
				},
			},
		},
		{
			name: "three ranges with different gaps",
			items: []testRange{
				{name: "r1", start: at(8, 0), end: at(9, 0)},
				{name: "r2", start: at(10, 0), end: at(11, 0)},
				{name: "r3", start: at(13, 0), end: at(14, 0)},
			},
			expected: []struct {
				before     time.Duration
				after      time.Duration
				searchFrom time.Time
				searchTo   time.Time
			}{
				{
					before:     0,
					after:      42 * time.Minute,
					searchFrom: at(8, 0),
					searchTo:   at(9, 42),
				},
				{
					before:     18 * time.Minute,
					after:      time.Hour,
					searchFrom: at(9, 42),
					searchTo:   at(12, 0),
				},
				{
					before:     36 * time.Minute,
					after:      0,
					searchFrom: at(12, 24),
					searchTo:   at(14, 0),
				},
			},
		},
		{
			name: "adjacent ranges produce no tolerance",
			items: []testRange{
				{name: "r1", start: at(9, 0), end: at(10, 0)},
				{name: "r2", start: at(10, 0), end: at(11, 0)},
			},
			expected: []struct {
				before     time.Duration
				after      time.Duration
				searchFrom time.Time
				searchTo   time.Time
			}{
				{
					before:     0,
					after:      0,
					searchFrom: at(9, 0),
					searchTo:   at(10, 0),
				},
				{
					before:     0,
					after:      0,
					searchFrom: at(10, 0),
					searchTo:   at(11, 0),
				},
			},
		},
		{
			name: "overlapping ranges produce no tolerance",
			items: []testRange{
				{name: "r1", start: at(9, 0), end: at(11, 0)},
				{name: "r2", start: at(10, 30), end: at(12, 0)},
			},
			expected: []struct {
				before     time.Duration
				after      time.Duration
				searchFrom time.Time
				searchTo   time.Time
			}{
				{
					before:     0,
					after:      0,
					searchFrom: at(9, 0),
					searchTo:   at(11, 0),
				},
				{
					before:     0,
					after:      0,
					searchFrom: at(10, 30),
					searchTo:   at(12, 0),
				},
			},
		},
		{
			name: "single item has no tolerance",
			items: []testRange{
				{name: "r1", start: at(9, 0), end: at(10, 0)},
			},
			expected: []struct {
				before     time.Duration
				after      time.Duration
				searchFrom time.Time
				searchTo   time.Time
			}{
				{
					before:     0,
					after:      0,
					searchFrom: at(9, 0),
					searchTo:   at(10, 0),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := CalculateSearchTolerances(tt.items, config)

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d results, got %d", len(tt.expected), len(result))
			}

			for i := range result {
				got := result[i]
				expected := tt.expected[i]

				if got.Tolerance.BeforeStart != expected.before {
					t.Fatalf(
						"item %d expected before tolerance %s, got %s",
						i,
						expected.before,
						got.Tolerance.BeforeStart,
					)
				}

				if got.Tolerance.AfterEnd != expected.after {
					t.Fatalf(
						"item %d expected after tolerance %s, got %s",
						i,
						expected.after,
						got.Tolerance.AfterEnd,
					)
				}

				searchFrom := got.Item.StartTime().Add(-got.Tolerance.BeforeStart)
				searchTo := got.Item.EndTime().Add(got.Tolerance.AfterEnd)

				if !searchFrom.Equal(expected.searchFrom) {
					t.Fatalf(
						"item %d expected searchFrom %s, got %s",
						i,
						expected.searchFrom.Format("15:04"),
						searchFrom.Format("15:04"),
					)
				}

				if !searchTo.Equal(expected.searchTo) {
					t.Fatalf(
						"item %d expected searchTo %s, got %s",
						i,
						expected.searchTo.Format("15:04"),
						searchTo.Format("15:04"),
					)
				}
			}
		})
	}
}

func TestCalculateSearchTolerances_CustomMaxValues(t *testing.T) {
	items := []testRange{
		{name: "r1", start: at(8, 0), end: at(9, 0)},
		{name: "r2", start: at(15, 0), end: at(16, 0)},
	}

	result, _ := CalculateSearchTolerances(items, ToleranceConfig{
		BeforePercent: 0.30,
		AfterPercent:  0.70,
		MaxBefore:     45 * time.Minute,
		MaxAfter:      90 * time.Minute,
	})

	searchFromR2 := result[1].Item.StartTime().Add(-result[1].Tolerance.BeforeStart)
	searchToR1 := result[0].Item.EndTime().Add(result[0].Tolerance.AfterEnd)

	if result[0].Tolerance.AfterEnd != 90*time.Minute {
		t.Fatalf("expected r1 after tolerance 90m, got %s", result[0].Tolerance.AfterEnd)
	}

	if searchToR1 != at(10, 30) {
		t.Fatalf("expected r1 searchTo 10:30, got %s", searchToR1.Format("15:04"))
	}

	if result[1].Tolerance.BeforeStart != 45*time.Minute {
		t.Fatalf("expected r2 before tolerance 45m, got %s", result[1].Tolerance.BeforeStart)
	}

	if searchFromR2 != at(14, 15) {
		t.Fatalf("expected r2 searchFrom 14:15, got %s", searchFromR2.Format("15:04"))
	}
}

func TestCalculateSearchTolerances_EmptyInput(t *testing.T) {
	result, _ := CalculateSearchTolerances[testRange](nil, ToleranceConfig{
		BeforePercent: 0.30,
		AfterPercent:  0.70,
		MaxBefore:     time.Hour,
		MaxAfter:      time.Hour,
	})

	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d", len(result))
	}
}

func TestToleranceConfigValidate_ReturnsError(t *testing.T) {
	tests := []struct {
		name        string
		config      ToleranceConfig
		expectedErr error
	}{
		{
			name: "negative before percent",
			config: ToleranceConfig{
				BeforePercent: -0.1,
				AfterPercent:  0.7,
				MaxBefore:     time.Hour,
				MaxAfter:      time.Hour,
			},
			expectedErr: ErrNegativePercent,
		},
		{
			name: "negative after percent",
			config: ToleranceConfig{
				BeforePercent: 0.3,
				AfterPercent:  -0.1,
				MaxBefore:     time.Hour,
				MaxAfter:      time.Hour,
			},
			expectedErr: ErrNegativePercent,
		},
		{
			name: "percent total greater than one",
			config: ToleranceConfig{
				BeforePercent: 0.6,
				AfterPercent:  0.5,
				MaxBefore:     time.Hour,
				MaxAfter:      time.Hour,
			},
			expectedErr: ErrInvalidPercentTotal,
		},
		{
			name: "negative max before",
			config: ToleranceConfig{
				BeforePercent: 0.3,
				AfterPercent:  0.7,
				MaxBefore:     -time.Minute,
				MaxAfter:      time.Hour,
			},
			expectedErr: ErrNegativeMax,
		},
		{
			name: "negative max after",
			config: ToleranceConfig{
				BeforePercent: 0.3,
				AfterPercent:  0.7,
				MaxBefore:     time.Hour,
				MaxAfter:      -time.Minute,
			},
			expectedErr: ErrNegativeMax,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestCalculateSearchTolerances_ReturnsErrorForInvalidConfig(t *testing.T) {
	items := []testRange{
		{name: "r1", start: at(9, 0), end: at(10, 0)},
		{name: "r2", start: at(12, 0), end: at(13, 0)},
	}

	result, err := CalculateSearchTolerances(items, ToleranceConfig{
		BeforePercent: 0.8,
		AfterPercent:  0.7,
		MaxBefore:     time.Hour,
		MaxAfter:      time.Hour,
	})

	if !errors.Is(err, ErrInvalidPercentTotal) {
		t.Fatalf("expected error %v, got %v", ErrInvalidPercentTotal, err)
	}

	if result != nil {
		t.Fatalf("expected nil result when config is invalid, got %#v", result)
	}
}
