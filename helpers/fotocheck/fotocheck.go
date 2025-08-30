package fotocheck

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"log/slog"
	"strings"
	"text/template"

	"github.com/signintech/gopdf"
)

//go:embed Futura_Std_Heavy.ttf
var futuraFont []byte

type FotocheckBuilder struct {
	base *template.Template
}

func NewFotocheckBuilder() *FotocheckBuilder {
	return &FotocheckBuilder{
		base: template.New("base"),
	}
}

func (b *FotocheckBuilder) fastStdHash(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum64())
}

func (b *FotocheckBuilder) renderTemplate(tplText string, data any) (string, error) {
	tplID := b.fastStdHash(tplText)

	tpl := b.base.Lookup(tplID)
	if tpl == nil {
		var err error
		tpl, err = b.base.New(tplID).Parse(tplText)
		if err != nil {
			return "", fmt.Errorf("failed to parse template: %w", err)
		}
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.String(), nil
}

func (b *FotocheckBuilder) totalPages(itemCount, rowsPerPage, colsPerPage int) int {
	if rowsPerPage == 0 || colsPerPage == 0 {
		return 0
	}
	itemsPerPage := rowsPerPage * colsPerPage
	return (itemCount + itemsPerPage - 1) / itemsPerPage
}

func (b *FotocheckBuilder) BuildPdf(ctx context.Context, data *FotocheckData) ([]byte, error) {
	if data.HeightMM == 0 || data.WidthMM == 0 {
		return nil, errors.New("invalid dimensions: height and width must be greater than 0")
	}
	if len(data.Items) == 0 {
		return nil, errors.New("no items provided to generate document")
	}
	if len(data.BackgroundJPG) == 0 {
		return nil, errors.New("background image (JPG) is required")
	}
	pdf := &gopdf.GoPdf{}
	defer pdf.Close()
	pdf.Start(gopdf.Config{
		Unit:     gopdf.UnitMM,
		PageSize: *gopdf.PageSizeA4,
	})
	pdf.AddTTFFontData("futura", futuraFont)
	pdf.SetFont("futura", "", 14)

	rows, err := XPositions(data.WidthMM)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	cols, err := YPositions(data.HeightMM)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	pages := b.totalPages(len(data.Items), len(rows), len(cols))
	if pages == 0 {
		return nil, errors.New("invalid layout: calculated total pages is 0")
	}

	b.buildPdf(pdf, data, pages, cols, rows)

	var doc bytes.Buffer
	if _, err := pdf.WriteTo(&doc); err != nil {
		slog.Error("failed to write PDF to buffer", "error", err)
		return nil, err
	}
	return doc.Bytes(), nil
}

func (b *FotocheckBuilder) addImage(pdf *gopdf.GoPdf, rect *gopdf.Rect, x, y float64, img gopdf.ImageHolder) error {
	if err := pdf.ImageByHolder(img, x, y, rect); err != nil {
		slog.Error("failed to add image to PDF", "error", err)
		return fmt.Errorf("failed to add image to PDF: %w", err)
	}
	return nil
}

func (b *FotocheckBuilder) buildPdf(pdf *gopdf.GoPdf, data *FotocheckData, pages int, cols []float64, rows []float64) error {
	backgroundHolder, err := gopdf.ImageHolderByBytes(data.BackgroundJPG)
	if err != nil {
		slog.Error("failed to create background image holder", "error", err)
		return fmt.Errorf("failed to create template image holder: %w", err)
	}

	var idx = 0
	var total = len(data.Items)
MAIN_LOOP:
	for range pages {
		pdf.AddPage()
		for _, col := range cols {
			for _, row := range rows {
				rect := &gopdf.Rect{W: data.WidthMM, H: data.HeightMM}
				if err := b.addImage(pdf, rect, row, col, backgroundHolder); err != nil {
					return nil
				}
				var item = data.Items[idx]
				for _, elm := range item.Elements {
					switch v := elm.(type) {
					case FotocheckText:
						text, err := b.renderTemplate(v.Text, item.Data)
						if err != nil {
							return fmt.Errorf("failed to render template at index %d: %w", idx, err)
						}
						var x = row + v.X*data.WidthMM
						var y = col + v.Y*data.HeightMM
						var fontsize = 10.0
						if v.FontSize > 0 {
							fontsize = v.FontSize
						}
						pdf.SetXY(x, y)
						if err := pdf.SetFontSize(fontsize); err != nil {
							return fmt.Errorf("failed to set font size at index %d: %w", idx, err)
						}
						if !v.Color.IsZero() {
							pdf.SetTextColor(v.Color.R, v.Color.G, v.Color.R)
						} else {
							pdf.SetTextColor(0, 0, 0)
						}
						if err := pdf.Text(text); err != nil {
							return fmt.Errorf("failed to write text at index %d: %w", idx, err)
						}

					case FotocheckImage:
						var x = row + v.X*data.WidthMM
						var y = col + v.Y*data.HeightMM
						var width = v.Width * data.WidthMM
						var height = v.Height * data.HeightMM
						var imageBytes = v.Bytes
						if v.Circle {
							cropped, err := circularCropBytes(imageBytes)
							if err != nil {
								slog.Warn("failed to apply circular crop", "error", err)
							} else {
								imageBytes = cropped
							}
						}
						imageHolder, err := gopdf.ImageHolderByBytes(imageBytes)
						if err != nil {
							slog.Error("failed to create image holder", "index", idx, "error", err)
							return fmt.Errorf("failed to create image holder at index %d: %w", idx, err)
						}
						imageRect := &gopdf.Rect{W: width, H: height}
						imageOpts := gopdf.ImageOptions{
							X:           x,
							Y:           y,
							Rect:        imageRect,
							DegreeAngle: v.Rotation,
						}
						if err := pdf.ImageByHolderWithOptions(imageHolder, imageOpts); err != nil {
							slog.Error("failed to insert image", "index", idx, "error", err)
							return fmt.Errorf("failed to insert image at index %d: %w", idx, err)
						}

					case FotocheckBarcode:
						var x = row + v.X*data.WidthMM
						var y = col + v.Y*data.HeightMM
						var width = v.Width * data.WidthMM
						var height = v.Height * data.HeightMM

						if strings.TrimSpace(v.Value) != "" {
							text, err := b.renderTemplate(strings.TrimSpace(v.Value), item.Data)
							if err != nil {
								return fmt.Errorf("failed to render template at index %d: %w", idx, err)
							}

							var barcodeImg io.Reader
							if v.Type == BarcodeTypeCode128 {
								barcodeImg, err = generateCode128Image(text)
							} else {
								barcodeImg, err = generateQRImage(text)
							}
							if err != nil {
								slog.Error("failed to generate barcode image", "index", idx, "type", v.Type, "error", err)
								return fmt.Errorf("failed to generate barcode image at index %d: %w", idx, err)
							}

							barcodeHolder, err := gopdf.ImageHolderByReader(barcodeImg)
							if err != nil {
								slog.Error("failed to create image holder from barcode", "index", idx, "type", v.Type, "error", err)
								return fmt.Errorf("failed to create image holder at index %d: %w", idx, err)
							}

							imageRect := &gopdf.Rect{W: width, H: height}
							imageOpts := gopdf.ImageOptions{
								X:           x,
								Y:           y,
								Rect:        imageRect,
								DegreeAngle: v.Rotation,
							}

							if err := pdf.ImageByHolderWithOptions(barcodeHolder, imageOpts); err != nil {
								slog.Error("failed to insert barcode image", "index", idx, "type", v.Type, "error", err)
								return fmt.Errorf("failed to insert barcode image at index %d: %w", idx, err)
							}
						}
					}
				}
				idx++
				if idx == total {
					break MAIN_LOOP
				}
			}
		}
	}
	return nil
}
