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

type ParagraphBoxAlign uint8

const (
	ParagraphBoxAlignLeft ParagraphBoxAlign = iota
	ParagraphBoxAlignCenter
	ParagraphBoxAlignRight
)

type FotocheckParagraphBox struct {
	FotocheckText
	Align       ParagraphBoxAlign // left=0, center=1, right=2
	Width       float64
	LineSpacing float64
	ShowBorder  bool
}

type FotocheckImage struct {
	Bytes []byte
	Position
	Size
	Rotation float64 // in degrees
	Circle   bool
}

// FotocheckLine draws a line inside a fotocheck. Start and End coordinates are
// relative to the fotocheck dimensions; LineWidth is measured in millimeters.
type FotocheckLine struct {
	Start     Position
	End       Position
	LineWidth float64
	Color     Color
	Style     LineStyle
}

type LineStyle uint8

const (
	LineStyleSolid LineStyle = iota
	LineStyleDashed
	LineStyleDotted
)

type BarcodeType uint8

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
var _ FotocheckElement = (*FotocheckParagraphBox)(nil)
var _ FotocheckElement = (*FotocheckLine)(nil)

func (FotocheckImage) isFotocheckElement()   {}
func (FotocheckText) isFotocheckElement()    {}
func (FotocheckBarcode) isFotocheckElement() {}
func (FotocheckLine) isFotocheckElement()    {}

type Fotocheck struct {
	Data     any
	Elements []FotocheckElement
}

type FotocheckData struct {
	PageWidthMM   float64
	PageHeightMM  float64
	MinGapX       float64
	MinGapY       float64
	PadStartX     float64
	PadEndX       float64
	PadStartY     float64
	PadEndY       float64
	WidthMM       float64
	HeightMM      float64
	BackgroundJPG []byte
	Items         []Fotocheck
}

var ErrImageType = errors.New("invalid image type: only JPEG format is supported")
