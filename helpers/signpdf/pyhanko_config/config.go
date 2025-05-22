package pyhankoconfig

import (
	"bytes"
	_ "embed"
	"html/template"
)

//go:embed pyhanko_unsigned.yml
var configWithoutImage string

//go:embed pyhanko_signed_with_img.yml
var configWithImage string

func RenderConfigWithoutImage() ([]byte, error) {
	return []byte(configWithoutImage), nil
}

func RenderConfigWithImage(backgroundPath string) ([]byte, error) {
	tmpl, err := template.New("withImage").Parse(configWithImage)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{
		"background": backgroundPath,
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
