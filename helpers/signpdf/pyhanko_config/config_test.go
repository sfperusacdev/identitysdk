package pyhankoconfig

import (
	"bytes"
	"os"
	"testing"
)

func TestBuildStampTextStripsControlChars(t *testing.T) {
	got := buildStampText("  ACME\u0091 Corp  ")
	want := "ACME Corp\n%(ts)s"

	if got != want {
		t.Fatalf("buildStampText() = %q, want %q", got, want)
	}
}

func TestRenderConfigWithoutImageQuotesStampText(t *testing.T) {
	config, err := RenderConfigWithoutImage("ACME\u0091 Corp")
	if err != nil {
		t.Fatalf("RenderConfigWithoutImage() error = %v", err)
	}

	if bytes.Contains(config, []byte{0x91}) {
		t.Fatalf("rendered config contains raw control byte 0x91: %q", config)
	}

	want := []byte(`stamp-text: "ACME Corp\n%(ts)s"`)
	if !bytes.Contains(config, want) {
		t.Fatalf("rendered config missing stamp text %q in %q", want, config)
	}
}

func TestRenderConfigWithImageQuotesBackgroundAndSanitizesText(t *testing.T) {
	bgFile := t.TempDir() + "/boleta img:1.png"
	if err := os.WriteFile(bgFile, []byte("img"), 0o600); err != nil {
		t.Fatalf("failed to create temp background file: %v", err)
	}

	config, err := RenderConfigWithImage("  ACME\u0091 Corp  ", bgFile)
	if err != nil {
		t.Fatalf("RenderConfigWithImage() error = %v", err)
	}

	if bytes.Contains(config, []byte{0x91}) {
		t.Fatalf("rendered config contains raw control byte 0x91: %q", config)
	}

	if !bytes.Contains(config, []byte(`background: `+"\""+bgFile+"\"")) {
		t.Fatalf("rendered config missing quoted background in %q", config)
	}

	if !bytes.Contains(config, []byte(`stamp-text: "ACME Corp\n%(ts)s"`)) {
		t.Fatalf("rendered config missing quoted stamp text in %q", config)
	}
}
