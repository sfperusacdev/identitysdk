package pyhankoconfig

import (
	"bytes"
	_ "embed"
	"html/template"
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
		"stamp-text": text,
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func RenderConfigWithImage(text string, backgroundPath string) ([]byte, error) {
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
		"stamp-text": text,
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
