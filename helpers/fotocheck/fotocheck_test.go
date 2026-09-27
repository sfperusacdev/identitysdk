package fotocheck

import (
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
