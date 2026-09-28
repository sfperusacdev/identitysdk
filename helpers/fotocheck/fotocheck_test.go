package fotocheck

import (
	"bytes"
	"context"
	"testing"
)

func TestBuildPdfNormalizesPageAndGapDefaults(t *testing.T) {
	tests := []struct {
		name           string
		pageWidth      float64
		pageHeight     float64
		minGapX        float64
		minGapY        float64
		wantPageWidth  float64
		wantPageHeight float64
		wantMinGapX    float64
		wantMinGapY    float64
	}{
		{
			name:           "zero dimensions and gaps",
			wantPageWidth:  210,
			wantPageHeight: 297,
		},
		{
			name:           "negative dimensions and gap normalize while positive values remain",
			pageWidth:      -1,
			pageHeight:     -1,
			minGapX:        -2,
			minGapY:        3,
			wantPageWidth:  210,
			wantPageHeight: 297,
			wantMinGapX:    0,
			wantMinGapY:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &FotocheckData{
				PageWidthMM:  tt.pageWidth,
				PageHeightMM: tt.pageHeight,
				MinGapX:      tt.minGapX,
				MinGapY:      tt.minGapY,
				WidthMM:      100,
				HeightMM:     50,
			}

			_, err := NewFotocheckBuilder().BuildPdf(context.Background(), data)
			if err == nil {
				t.Fatal("BuildPdf() error = nil, want error for empty items")
			}

			if data.PageWidthMM != tt.wantPageWidth {
				t.Errorf("PageWidthMM = %v, want %v", data.PageWidthMM, tt.wantPageWidth)
			}
			if data.PageHeightMM != tt.wantPageHeight {
				t.Errorf("PageHeightMM = %v, want %v", data.PageHeightMM, tt.wantPageHeight)
			}
			if data.MinGapX != tt.wantMinGapX {
				t.Errorf("MinGapX = %v, want %v", data.MinGapX, tt.wantMinGapX)
			}
			if data.MinGapY != tt.wantMinGapY {
				t.Errorf("MinGapY = %v, want %v", data.MinGapY, tt.wantMinGapY)
			}
		})
	}
}

func TestBuildPdfWithLine(t *testing.T) {
	data := &FotocheckData{
		WidthMM:  100,
		HeightMM: 50,
		Items: []Fotocheck{{
			Elements: []FotocheckElement{
				FotocheckLine{
					Start:     Position{X: 0.1, Y: 0.2},
					End:       Position{X: 0.9, Y: 0.8},
					LineWidth: 0.5,
					Color:     Color{R: 255, G: 0, B: 0},
				},
				FotocheckLine{
					Start: Position{X: 0.1, Y: 0.5},
					End:   Position{X: 0.9, Y: 0.5},
					Style: LineStyleDashed,
				},
				FotocheckLine{
					Start: Position{X: 0.1, Y: 0.8},
					End:   Position{X: 0.9, Y: 0.8},
					Style: LineStyleDotted,
				},
			},
		}},
	}

	pdf, err := NewFotocheckBuilder().BuildPdf(context.Background(), data)
	if err != nil {
		t.Fatalf("BuildPdf() error = %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatal("BuildPdf() did not return a PDF document")
	}
}
