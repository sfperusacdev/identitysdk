package fotocheck

import (
	"errors"
	"math"
	"testing"
)

func TestSegments(t *testing.T) {
	const a4Width float64 = 210
	const minGapX float64 = 15
	tests := []struct {
		name  string
		width float64
		cols  []float64
		err   error
	}{
		{"T1", 210, nil, ErrElementTooWide},
		{"T1-1", a4Width - minGapX, []float64{math.Trunc(minGapX / 2)}, nil},
		{"T2", 200, nil, ErrElementTooWide},
		{"T2", 195, []float64{7}, nil},
		{"T3", 100, []float64{55}, nil},
		{"T4", 90, []float64{10, 110}, nil},
		{"T5", 0, []float64{}, ErrElementTooNarrow},
		{"T6", -1, []float64{}, ErrElementTooNarrow},
		{"T7", minElementWidth - 1, []float64{}, ErrElementTooNarrow},
		{"T8", 50, []float64{15, 80, 145}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cols, err := segments(a4Width, tt.width, minGapX)
			if !errors.Is(err, tt.err) {
				t.Errorf("expected err: %v, got: %v", tt.err, err)
			}
			if len(cols) != len(tt.cols) {
				t.Errorf("expected len=%d, got len=%d", len(tt.cols), len(cols))
				t.Logf("result cols: %v", cols)
				return
			}
			for i, val := range tt.cols {
				if val != cols[i] {
					t.Errorf("expected [i=%d] %f, got %f", i, val, cols[i])
					t.Logf("full result cols: %v", cols)
				}
			}
		})
	}
}
