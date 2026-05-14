package signpdf

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/rs/xid"
	"github.com/sfperusacdev/identitysdk"
	pyhankoconfig "github.com/sfperusacdev/identitysdk/helpers/signpdf/pyhanko_config"
	"github.com/user0608/ifdevmode"
)

type PyhankoPDFSigner struct {
	pyhankoBin string
	inspector  PDFInspector
}

var _ PDFSigner = (*PyhankoPDFSigner)(nil)

func NewPyhankoPDFSigner(inspector PDFInspector) *PyhankoPDFSigner {
	return NewPyhankoPDFSignerWithPyhankoPath(inspector, "pyhanko")
}
func NewPyhankoPDFSignerWithPyhankoPath(inspector PDFInspector, pyhankoBin string) *PyhankoPDFSigner {
	return &PyhankoPDFSigner{
		pyhankoBin: pyhankoBin,
		inspector:  inspector,
	}
}
func (py *PyhankoPDFSigner) prepareConfigFile(box *SignBox, text string) (string, func() error, error) {
	var configData []byte
	var err error
	var cleanupImg func() error
	if box != nil && box.ImageBackground != nil {
		var imgPath string

		imgPath, cleanupImg, err = py.storeTempFile(box.ImageBackground)
		if err != nil {
			return "", nil, fmt.Errorf("failed to store background image: %w", err)
		}

		configData, err = pyhankoconfig.RenderConfigWithImage(text, imgPath)
		if err != nil {
			return "", nil, fmt.Errorf("failed to render config with image: %w", err)
		}
	} else {
		configData, err = pyhankoconfig.RenderConfigWithoutImage(text)
		if err != nil {
			return "", nil, fmt.Errorf("failed to render config without image: %w", err)
		}
	}

	path, cleanupConfig, err := py.storeTempFile(configData)
	if err != nil {
		return "", nil, fmt.Errorf("failed to store config file: %w", err)
	}

	cleanupFunc := func() error {
		if cleanupImg != nil {
			_ = cleanupImg()
		}
		if cleanupConfig != nil {
			_ = cleanupConfig()
		}
		return nil
	}

	return path, cleanupFunc, nil
}

func (s *PyhankoPDFSigner) storeKeyAndCertFiles(keyPEM, certPEM []byte) (keyPath string, certPath string, cleanup func() error, err error) {
	tempDir, err := os.MkdirTemp("", "cert_files_")
	if err != nil {
		slog.Error("failed to create temporary directory", "error", err)
		return "", "", nil, err
	}

	keyPath = filepath.Join(tempDir, "key.pem")
	certPath = filepath.Join(tempDir, "cert.pem")

	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil { // Permisos 0600 para el archivo de la clave.
		slog.Error("failed to write key file", "error", err, "path", keyPath)
		return "", "", nil, err
	}

	if err := os.WriteFile(certPath, certPEM, 0644); err != nil { // Permisos 0644 para el archivo del certificado.
		slog.Error("failed to write cert file", "error", err, "path", certPath)
		return "", "", nil, err
	}

	cleanupFunc := func() error {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			slog.Error("failed to remove temporary directory", "error", removeErr, "path", tempDir)
			return removeErr
		}
		return nil
	}

	return keyPath, certPath, cleanupFunc, nil
}

func (s *PyhankoPDFSigner) storeTempFile(data []byte) (string, func() error, error) {
	mime := mimetype.Detect(data)
	ext := mime.Extension()

	tempFile, err := os.CreateTemp("", "tempfile_*"+ext)
	if err != nil {
		slog.Error("failed to create temporary file", "error", err)
		return "", nil, err
	}
	defer tempFile.Close()

	if _, err := tempFile.Write(data); err != nil {
		slog.Error("failed to write to temporary file", "error", err, "path", tempFile.Name())
		return "", nil, err
	}

	tempFilePath := tempFile.Name()

	cleanupFunc := func() error {
		if err := os.Remove(tempFilePath); err != nil {
			slog.Error("failed to remove temporary file", "error", err, "path", tempFilePath)
			return err
		}
		return nil
	}

	return tempFilePath, cleanupFunc, nil
}

func (s *PyhankoPDFSigner) outputPath(id string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("signed_%s.pdf", id))
}

func (py *PyhankoPDFSigner) buildSignField(signName string, box SignBox, widthPt, heightPt float64) string {
	x1 := int(box.LeftPercent * widthPt)
	x2 := int((box.LeftPercent + box.WidthPercent) * widthPt)

	y2 := int((1 - box.TopPercent) * heightPt)
	y1 := y2 - int(box.HeightPercent*heightPt)

	return fmt.Sprintf("%d/%d,%d,%d,%d/%s", box.Page, x1, y1, x2, y2, signName)
}

func (py *PyhankoPDFSigner) ensureBoxDefaults(box *SignBox) {
	if box == nil {
		return
	}

	clamp := func(val, min, max float64) float64 {
		if val < min {
			return min
		}
		if val > max {
			return max
		}
		return val
	}

	box.TopPercent = clamp(box.TopPercent, 0, 1)
	box.LeftPercent = clamp(box.LeftPercent, 0, 1)

	if box.WidthPercent < 0.2 {
		box.WidthPercent = 0.2
	}

	if box.HeightPercent < 0.07 {
		box.HeightPercent = 0.07
	}
}

func (s *PyhankoPDFSigner) extractCommonName(certPEM string) string {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil || block.Type != "CERTIFICATE" {
		slog.Warn("Failed to decode PEM block or invalid certificate type")
		return ""
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		slog.Warn("Failed to parse certificate", "error", err)
		return ""
	}
	return cert.Subject.CommonName
}

func (py *PyhankoPDFSigner) Sign(
	ctx context.Context,
	signName string,
	pdfData []byte,
	certPEM, keyPEM string,
	box *SignBox,
) (*SignedPDFResult, error) {
	if !mimetype.Detect(pdfData).Is("application/pdf") {
		return nil, errors.New("invalid file type: expected application/pdf")
	}

	py.ensureBoxDefaults(box)

	keyPath, certPath, cleanUpKeys, err := py.storeKeyAndCertFiles([]byte(keyPEM), []byte(certPEM))
	if cleanUpKeys != nil && !ifdevmode.Yes() {
		defer cleanUpKeys()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to store key and certificate: %w", err)
	}

	configPath, configCleanup, err := py.prepareConfigFile(box, py.extractCommonName(certPEM))
	if configCleanup != nil && !ifdevmode.Yes() {
		defer configCleanup()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to prepare config file: %w", err)
	}

	inputPDFPath, inputCleanup, err := py.storeTempFile(pdfData)
	if inputCleanup != nil && !ifdevmode.Yes() {
		defer inputCleanup()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to store input PDF: %w", err)
	}

	id := xid.New().String()
	safeSignName := fmt.Sprintf("%s_%s", strings.ReplaceAll(signName, " ", "_"), id)

	fieldValue := fmt.Sprintf("1/0,0,0,0/%s", safeSignName)
	if box != nil {
		widthPt, heightPt, err := py.inspector.GetPageDimensionsFromPath(inputPDFPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get page dimensions: %w", err)
		}

		fieldValue = py.buildSignField(safeSignName, *box, widthPt, heightPt)
	}

	outputPath := py.outputPath(id)

	args := []string{
		"--verbose",
		"--config", configPath,
		"sign", "addsig",
		"--field", fieldValue,
		"pemder",
		"--key", keyPath,
		"--cert", certPath,
		inputPDFPath,
		outputPath,
		"--no-pass",
	}

	stdout, stderr, exitCode, err := py.runPyhanko(ctx, args...)
	if err != nil {
		details := strings.TrimSpace(stderr)
		if details == "" {
			details = strings.TrimSpace(stdout)
		}
		if details == "" {
			details = err.Error()
		}

		slog.Error(
			"pyHanko command failed",
			"bin", py.pyhankoBin,
			"args", args,
			"exit_code", exitCode,
			"error", err,
			"stdout", stdout,
			"stderr", stderr,
		)

		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, fmt.Errorf("pyHanko signing cancelled or timed out: %w: %s", ctxErr, details)
		}

		return nil, fmt.Errorf("failed to sign PDF with pyHanko: exit_code=%d: %s", exitCode, details)
	}

	signedPDF, err := os.ReadFile(outputPath)
	if err != nil {
		slog.Error("failed to read signed PDF", "path", outputPath, "error", err)
		return nil, fmt.Errorf("failed to read signed PDF: %w", err)
	}

	if err := os.Remove(outputPath); err != nil {
		slog.Error("failed to remove signed PDF file", "path", outputPath, "error", err)
	}

	return &SignedPDFResult{
		ID:                id,
		SingnedPDFContent: signedPDF,
	}, nil
}

func (py *PyhankoPDFSigner) runPyhanko(
	ctx context.Context,
	args ...string,
) (stdout string, stderr string, exitCode int, err error) {
	cmd := exec.CommandContext(ctx, py.pyhankoBin, args...)

	var tzStr = "TZ=America/Lima"
	loc, _ := identitysdk.Tz(ctx)
	if loc != nil {
		tzStr = fmt.Sprintf("TZ=%s", loc.String())
	}
	cmd.Env = append(os.Environ(), tzStr)

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()

	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	if err == nil {
		return stdout, stderr, 0, nil
	}

	exitCode = -1

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		return stdout, stderr, exitCode, ctxErr
	}

	return stdout, stderr, exitCode, err
}

func (py *PyhankoPDFSigner) SignFromFile(ctx context.Context, signName string, pdfPath, certPEM, keyPEM string, box *SignBox) (*SignedPDFResult, error) {
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		slog.Error("failed to read PDF file", "path", pdfPath, "error", err)
		return nil, err
	}
	return py.Sign(ctx, signName, pdfData, certPEM, keyPEM, box)
}
