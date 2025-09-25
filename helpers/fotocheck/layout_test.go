package fotocheck

import (
	"errors"
	"math"
	"testing"
)

func TestSegments(t *testing.T) {
	const a4Width float64 = 210
	const a4WidthFull float64 = 225
	const gap float64 = 15
	const padStart float64 = 5
	const padEnd float64 = 10

	tests := []struct {
		name  string
		width float64
		cols  []float64
		err   error
	}{
		{"T1", a4Width, []float64{padStart}, nil},
		{"T1-1", a4Width - gap, []float64{padStart}, nil},
		{"T2", 200, []float64{padStart}, nil},
		{"T2", (a4Width - gap) / 2, []float64{padStart, padStart + gap + (a4Width-gap)/2}, nil},
		{"T2", ((a4Width - gap) / 2) + 0.00001, []float64{padStart}, nil},
		{"T2", ((a4Width - gap) / 2) - 1, []float64{padStart, padStart + gap + 2 + (((a4Width - gap) / 2) - 1)}, nil},
		{"T2-1", 195, []float64{padStart}, nil},
		{"T3", 100, []float64{padStart}, nil},
		{"T4", 90, []float64{padStart, 120 + padStart}, nil},
		{"T5", 0, nil, ErrInvalidWidth},
		{"T6", -1, nil, ErrInvalidWidth},
		{"T7", 1, []float64{
			padStart + 0,
			padStart + 16.076923076923077,
			padStart + 32.15384615384615,
			padStart + 48.230769230769226,
			padStart + 64.3076923076923,
			padStart + 80.38461538461539,
			padStart + 96.46153846153847,
			padStart + 112.53846153846155,
			padStart + 128.6153846153846,
			padStart + 144.69230769230768,
			padStart + 160.76923076923075,
			padStart + 176.8461538461538,
			padStart + 192.92307692307688,
			padStart + 208.99999999999994,
		}, nil},
		{"T8", 50, []float64{padStart, padStart + 80, padStart + 160}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cols, err := segments(a4WidthFull, tt.width, gap, padStart, padEnd)
			if !errors.Is(err, tt.err) {
				t.Errorf("expected err: %v, got: %v", tt.err, err)
			}
			if tt.err != nil {
				return
			}
			if len(cols) != len(tt.cols) {
				t.Errorf("expected len=%d, got len=%d", len(tt.cols), len(cols))
				t.Logf("result cols: %v", cols)
				return
			}
			for i, val := range tt.cols {
				if math.Trunc(val) != math.Trunc(cols[i]) {
					t.Errorf("expected [i=%d] %f, got %f", i, val, cols[i])
					t.Logf("full result cols: %v", cols)
				}
			}
		})
	}
}

func TestSegmentsErrors(t *testing.T) {
	tests := []struct {
		name     string
		length   float64
		width    float64
		gap      float64
		padStart float64
		padEnd   float64
		err      error
	}{
		{"E1", 0, 10, 15, 0, 0, ErrInvalidLength},
		{"E2", 210, 10, -1, 0, 0, ErrInvalidGap},
		{"E3", 210, 10, 15, -1, 0, ErrInvalidPadding},
		{"E4", 210, 10, 15, 0, -5, ErrInvalidPadding},
		{"E5", 50, 60, 15, 0, 0, ErrNotEnoughSpace},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := segments(tt.length, tt.width, tt.gap, tt.padStart, tt.padEnd)
			if !errors.Is(err, tt.err) {
				t.Errorf("expected err: %v, got: %v", tt.err, err)
			}
		})
	}
}
