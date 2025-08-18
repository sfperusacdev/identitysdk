package fotocheck

import "errors"

type Position struct {
	X float64
	Y float64
}

type Size struct {
	Width  float64
	Height float64
}

type Color struct {
	R uint8
	G uint8
	B uint8
}

func (c *Color) IsZero() bool { return c.R == 0 && c.G == 0 && c.B == 0 }

type FotocheckText struct {
	Text     string
	FontSize float64 // default 10
	Color    Color
	Position
}

type FotocheckImage struct {
	Bytes []byte
	Position
	Size
	Rotation float64 // in degrees
}

type BarcodeType uint

const (
	BarcodeTypeQR BarcodeType = iota
	BarcodeTypeCode128
)

type FotocheckBarcode struct {
	Value string
	Position
	Size
	Rotation float64
	Type     BarcodeType
}

// FotocheckElement represents any supported element type on the card.
type FotocheckElement interface {
	isFotocheckElement()
}

var _ FotocheckElement = (*FotocheckImage)(nil)
var _ FotocheckElement = (*FotocheckText)(nil)
var _ FotocheckElement = (*FotocheckBarcode)(nil)

func (FotocheckImage) isFotocheckElement()   {}
func (FotocheckText) isFotocheckElement()    {}
func (FotocheckBarcode) isFotocheckElement() {}

type Fotocheck struct {
	Data     any
	Elements []FotocheckElement
}

type FotocheckData struct {
	WidthMM       float64
	HeightMM      float64
	BackgroundJPG []byte
	Items         []Fotocheck
}

var ErrImageType = errors.New("invalid image type: only JPEG format is supported")
