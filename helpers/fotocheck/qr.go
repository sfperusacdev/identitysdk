package fotocheck

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io"
	"log/slog"

	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/qr"
)

func generateCode128Image(content string) (io.Reader, error) {
	barcode, err := code128.Encode(content)
	if err != nil {
		slog.Error("failed to encode Code128", "error", err)
		return nil, fmt.Errorf("failed to encode Code128: %w", err)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, barcode, &jpeg.Options{Quality: 100}); err != nil {
		slog.Error("failed to encode Code128 as JPEG", "error", err)
		return nil, fmt.Errorf("failed to encode Code128 as JPEG: %w", err)
	}

	return &buf, nil
}

func generateQRImage(content string) (io.Reader, error) {
	qrCode, err := qr.Encode(content, qr.Q, qr.Auto)
	if err != nil {
		slog.Error("failed to encode QR", "error", err)
		return nil, fmt.Errorf("failed to encode QR: %w", err)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, qrCode, &jpeg.Options{Quality: 100}); err != nil {
		slog.Error("failed to encode QR as JPEG", "error", err)
		return nil, fmt.Errorf("failed to encode QR as JPEG: %w", err)
	}

	return &buf, nil
}
