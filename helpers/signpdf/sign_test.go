package signpdf

import (
	"strings"
	"testing"
)

func TestBuildSignField(t *testing.T) {
	signer := &PyhankoPDFSigner{}
	box := SignBox{
		Page:          1,
		TopPercent:    0.874,
		LeftPercent:   0.684,
		WidthPercent:  0.198,
		HeightPercent: 0.07,
	}
	widthPt := 595.28  // PDF width in points
	heightPt := 841.89 // PDF height in points

	result := signer.buildSignField("John Doe", box, widthPt, heightPt)

	expected := "1/407,48,525,106/John_Doe"
	if result != expected {
		t.Errorf("unexpected result:\nexpected: %s\ngot: %s", expected, result)
	}

	if !strings.HasSuffix(result, "/John_Doe") {
		t.Errorf("signature name was not sanitized or appended correctly")
	}
}
