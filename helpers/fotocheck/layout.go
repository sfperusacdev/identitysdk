package fotocheck

import (
	"errors"
	"math"
)

var ErrInvalidLength = errors.New("length must be greater than 0")
var ErrInvalidWidth = errors.New("width must be greater than 0")
var ErrInvalidGap = errors.New("gap must be greater or equal to 0")
var ErrInvalidPadding = errors.New("padding must be greater or equal to 0")
var ErrNotEnoughSpace = errors.New("not enough space for elements with given paddings and gap")

// segments calcula posiciones X iniciales para elementos dentro de un contenedor.
// length: longitud total del contenedor (>=0)
// width: ancho de cada elemento (>0)
// gapMin: separación mínima entre elementos (>=0)
// padStart: espacio al inicio (>=0)
// padEnd: espacio al final (>=0)
func segments(length, width, gapMin, padStart, padEnd float64) ([]float64, error) {
	if length <= 0 {
		return nil, ErrInvalidLength
	}
	if width <= 0 {
		return nil, ErrInvalidWidth
	}
	if gapMin < 0 {
		return nil, ErrInvalidGap
	}
	if padStart < 0 || padEnd < 0 {
		return nil, ErrInvalidPadding
	}

	usable := length - padStart - padEnd
	if usable < width {
		return nil, ErrNotEnoughSpace
	}

	// número máximo de elementos posibles con el gap mínimo
	n := math.Floor((usable + gapMin) / (width + gapMin))
	if n <= 0 {
		return nil, ErrNotEnoughSpace
	}

	// recalcular gap real distribuyendo el espacio sobrante
	totalWidth := n*width + (n-1)*gapMin
	freeSpace := usable - totalWidth
	gap := gapMin
	if n > 1 {
		gap += freeSpace / float64(n-1)
	}

	pos := make([]float64, int(n))
	x := padStart
	for i := 0; i < int(n); i++ {
		pos[i] = x
		x += width + gap
	}
	return pos, nil
}
