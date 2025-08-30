package fotocheck

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
)

func circularCropBytes(jpgIn []byte) ([]byte, error) {
	src, _, err := image.Decode(bytes.NewReader(jpgIn))
	if err != nil {
		// asegúrate de registrar: _ "image/jpeg"
		if _, jerr := jpeg.Decode(bytes.NewReader(jpgIn)); jerr != nil {
			return nil, err
		}
	}

	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	diam := sw
	if sh < diam {
		diam = sh
	}

	dst := image.NewRGBA(image.Rect(0, 0, diam, diam))

	ox := (sw - diam) / 2
	oy := (sh - diam) / 2
	draw.Draw(dst, dst.Bounds(), src, image.Pt(ox, oy), draw.Src)

	r := diam / 2
	cx, cy := r, r
	rr := r * r
	for y := 0; y < diam; y++ {
		for x := 0; x < diam; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy > rr {
				i := dst.PixOffset(x, y)
				dst.Pix[i+3] = 0
			}
		}
	}

	var out bytes.Buffer
	if err := png.Encode(&out, dst); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
