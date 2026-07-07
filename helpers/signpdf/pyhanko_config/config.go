package pyhankoconfig

import (
	"bytes"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"text/template"
	"unicode"

	"github.com/gosimple/unidecode"
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
	text = buildStampText(text)
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{
		"stamp_text": text,
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func SanitizeYAMLText(s string) string {
	// Convierte Unicode a ASCII
	s = unidecode.Unidecode(s)

	// Elimina caracteres de control que YAML no acepta
	s = strings.Map(func(r rune) rune {
		switch {
		case r == '\n', r == '\r', r == '\t':
			return r
		case r < 0x20, r == 0x7F, (r >= 0x80 && r <= 0x9F):
			return -1
		default:
			return r
		}
	}, s)

	return s
}

func RenderConfigWithImage(text string, backgroundPath string) ([]byte, error) {
	text = SanitizeYAMLText(text)
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
	text = buildStampText(text)
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

func buildStampText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		text = "%(signer)s"
	}

	text = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) && !unicode.IsControl(r) {
			return r
		}
		return -1
	}, text)

	return fmt.Sprintf("%s\n%%(ts)s", text)
}
