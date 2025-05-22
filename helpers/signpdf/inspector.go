package signpdf

import (
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/gabriel-vasile/mimetype"
)

type PopplerPDFInspector struct {
	pageSizeRegex *regexp.Regexp
}

var _ PDFInspector = (*PopplerPDFInspector)(nil)

func NewPopplerPDFInspector() *PopplerPDFInspector {
	return &PopplerPDFInspector{
		pageSizeRegex: regexp.MustCompile(`Page size:\s+([\d.]+)\s+x\s+([\d.]+)`),
	}
}

func (p *PopplerPDFInspector) GetPageDimensionsFromPath(pdfPath string) (widthPt float64, heightPt float64, err error) {
	cmd := exec.Command("pdfinfo", pdfPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	matches := p.pageSizeRegex.FindStringSubmatch(string(output))
	if len(matches) != 3 {
		return 0, 0, errors.New("page size not found")
	}

	width, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, 0, err
	}
	height, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return 0, 0, err
	}

	return width, height, nil
}

func (p *PopplerPDFInspector) GetPageDimensionsFromBytes(data []byte) (float64, float64, error) {
	if !mimetype.Detect(data).Is("application/pdf") {
		return 0, 0, errors.New("invalid file type: expected application/pdf")
	}

	tmpFile, err := os.CreateTemp("", "pdf-inspect-*.pdf")
	if err != nil {
		return 0, 0, err
	}

	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(data); err != nil {
		return 0, 0, err
	}

	return p.GetPageDimensionsFromPath(tmpFile.Name())
}
