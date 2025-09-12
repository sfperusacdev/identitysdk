package pdfutils

import (
	"strings"

	"github.com/signintech/gopdf"
)

func wrapLines(pdf *gopdf.GoPdf, text string, w float64) ([]string, error) {
	lines := []string{}
	for _, rawLine := range strings.Split(text, "\n") {
		words := strings.Fields(rawLine)
		cur := ""
		breakLong := func(word string) ([]string, error) {
			var parts []string
			for len(word) > 0 {
				runes := []rune(word)
				i := 1
				for ; i <= len(runes); i++ {
					tw, err := pdf.MeasureTextWidth(string(runes[:i]))
					if err != nil {
						return nil, err
					}
					if tw > w {
						break
					}
				}
				if i == 1 {
					parts = append(parts, string(runes[:1]))
					word = string(runes[1:])
				} else if i > len(runes) {
					parts = append(parts, word)
					break
				} else {
					parts = append(parts, string(runes[:i-1]))
					word = string(runes[i-1:])
				}
			}
			return parts, nil
		}

		for _, word := range words {
			test := word
			if cur != "" {
				test = cur + " " + word
			}
			tw, err := pdf.MeasureTextWidth(test)
			if err != nil {
				return nil, err
			}
			if tw > w && cur != "" {
				lines = append(lines, cur)
				cur = word
			} else if tw > w {
				breaks, err := breakLong(word)
				if err != nil {
					return nil, err
				}
				for _, part := range breaks {
					if cur == "" {
						cur = part
					} else {
						lines = append(lines, cur)
						cur = part
					}
				}
			} else {
				cur = test
			}
		}
		if cur != "" {
			lines = append(lines, cur)
		}
	}
	if len(lines) == 0 {
		return []string{""}, nil
	}
	return lines, nil
}

func DrawTextInBox(pdf *gopdf.GoPdf, x, y, w float64, text string, align uint, lineSpacing float64, showBorder bool) error {
	lines, err := wrapLines(pdf, text, w)
	if err != nil {
		return err
	}
	lineH, err := pdf.MeasureCellHeightByText(text)
	if err != nil {
		return err
	}
	lineH += lineSpacing
	if showBorder {
		boxH := lineH * float64(len(lines))
		pdf.RectFromUpperLeftWithStyle(x, y, w, boxH, "D")
	}
	for i, line := range lines {
		tw, err := pdf.MeasureTextWidth(line)
		if err != nil {
			return err
		}
		tx := x
		switch align {
		case 2:
			tx = x + w - tw
		case 1:
			tx = x + (w-tw)/2
		}
		if tx < x {
			tx = x
		}
		pdf.SetX(tx)
		pdf.SetY(y + float64(i)*lineH)
		pdf.Cell(nil, line)
	}
	return nil
}
