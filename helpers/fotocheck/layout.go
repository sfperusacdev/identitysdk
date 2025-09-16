package fotocheck

import (
	"errors"
	"math"
)

var ErrElementTooWide = errors.New("element width exceeds container size")
var ErrElementTooNarrow = errors.New("element width is below minimum, min=10mm")
var ErrGapTooSmall = errors.New("gap between elements is below minimum required")

// segments receives a container length `length` and an element width `width`,
// and returns the starting X positions of as many elements as can fit,
// ensuring at least `gapMin` space between them.
//
// `length` must be greater than or equal to `width + gapMin`, otherwise an error is returned.
const minElementWidth float64 = 10 // Minimum element width in mm
func segments(length float64, width float64, gapMin float64) ([]float64, error) {
	if width < minElementWidth {
		return nil, ErrElementTooNarrow
	}

	n := math.Trunc(length / (width + gapMin))
	if n <= 0 {
		return nil, ErrElementTooWide
	}

	gap := math.Trunc((length - n*width) / (n + 1))
	if gap < math.Trunc(gapMin/2) {
		return nil, ErrGapTooSmall
	}

	pos := make([]float64, int(n))
	x := gap
	for i := 0; i < int(n); i++ {
		pos[i] = x
		x += width + gap
	}
	return pos, nil
}

// xPositions returns the horizontal starting positions for elements of width `w`,
// fitting as many as possible in an A4 page width, respecting minimum horizontal gap.
func XPositions(pageWidth, w, minGap float64) ([]float64, error) {
	return segments(pageWidth, w, minGap)
}

// yPositions returns the vertical starting positions for elements of height `h`,
// fitting as many as possible in an A4 page height, respecting minimum vertical gap.
func YPositions(pageHeight, h, minGap float64) ([]float64, error) {
	return segments(pageHeight, h, minGap)
}
