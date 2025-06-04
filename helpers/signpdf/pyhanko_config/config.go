package pyhankoconfig

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"strings"
)

//go:embed pyhanko_unsigned.yml
var configWithoutImage string

//go:embed pyhanko_signed_with_img.yml
var configWithImage string

func RenderConfigWithoutImage(text string) ([]byte, error) {
	tmpl, err := template.New("withoutImage").Parse(configWithoutImage)
	if err != nil {
		return nil, err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		text = "%(signer)s"
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{
		"stamp_text": text,
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func RenderConfigWithImage(text string, backgroundPath string) ([]byte, error) {
	info, err := os.Stat(backgroundPath)
	if err != nil {
		slog.Error("failed to stat background path", slog.String("path", backgroundPath), slog.Any("error", err))
		return nil, fmt.Errorf("background path error: %w", err)
	}
	if info.IsDir() {
		slog.Error("background path is a directory", slog.String("path", backgroundPath))
		return nil, fmt.Errorf("background path is a directory, not a file")
	}

	tmpl, err := template.New("withImage").Parse(configWithImage)
	if err != nil {
		return nil, err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		text = "%(signer)s"
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{
		"background": backgroundPath,
		"stamp_text": text,
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
