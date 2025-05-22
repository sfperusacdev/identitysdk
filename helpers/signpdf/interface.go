package signpdf

import "context"

type SignedPDFResult struct {
	ID                string
	SingnedPDFContent []byte
}

type SignBox struct {
	Page            int
	TopPercent      float64
	LeftPercent     float64
	WidthPercent    float64
	HeightPercent   float64
	ImageBackground []byte
}

type PDFSigner interface {
	// Sign signs a PDF provided as a byte slice using the given certificate and private key.
	// If box is nil, the signature will appear in the PDF as metadata without a visible stamp.
	// If box is provided, the signature will be placed visually at the specified coordinates.
	Sign(ctx context.Context, signName string, pdfData []byte, certPEM, keyPEM string, box *SignBox) (*SignedPDFResult, error)

	// SignFromFile signs a PDF located at the given file path using the specified certificate and private key.
	// If box is nil, the signature will appear in the PDF as metadata without a visible stamp.
	// If box is provided, the signature will be placed visually at the specified coordinates.
	SignFromFile(ctx context.Context, signName string, pdfPath, certPEM, keyPEM string, box *SignBox) (*SignedPDFResult, error)
}

type PDFInspector interface {
	GetPageDimensionsFromPath(pdfPath string) (widthPt float64, heightPt float64, err error)
	GetPageDimensionsFromBytes(data []byte) (widthPt float64, heightPt float64, err error)
}
