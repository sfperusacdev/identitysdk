package docxtopdf

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/sfperusacdev/identitysdk/services"
	"github.com/user0608/goones/errs"
)

type DocxTemplateToPdfService struct {
	bridge *services.ExternalBridgeService
}

func NewDocxTemplateToPdfService(bridge *services.ExternalBridgeService) *DocxTemplateToPdfService {
	return &DocxTemplateToPdfService{
		bridge: bridge,
	}
}

func (s *DocxTemplateToPdfService) resolveServiceURL(ctx context.Context) (string, error) {
	const variableName = "docx_to_pdf_service_url"

	getValidURL := func(value string) (string, bool) {
		if _, err := url.ParseRequestURI(value); err == nil {
			return value, true
		}
		return "", false
	}

	rawURL, err := s.bridge.ReadVariable(ctx, variableName)
	if err == nil {
		if validURL, ok := getValidURL(rawURL); ok {
			return validURL, nil
		}
	} else if !errors.Is(err, services.ErrVariableNotFound) {
		return "", err
	}

	rawURL, err = s.bridge.ReadVariableGlobal(ctx, variableName)
	if err != nil {
		if errors.Is(err, services.ErrVariableNotFound) {
			return "", errs.InternalErrorDirect("la variable `docx_to_pdf_service_url` no fue encontrada")
		}
		return "", err
	}

	if validURL, ok := getValidURL(rawURL); ok {
		return validURL, nil
	}

	return "", errs.InternalErrorDirect("la variable `docx_to_pdf_service_url` contiene un valor no válido")
}

const MIMEWordDocx = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

func (s *DocxTemplateToPdfService) GeneratePDF(ctx context.Context, templateBytes []byte, data any) ([]byte, error) {
	fileType := mimetype.Detect(templateBytes)
	if !fileType.Is(MIMEWordDocx) {
		return nil, errs.InternalErrorDirect("el template proporcionado al servicio de PDF no es un archivo .docx válido")
	}

	var payload = struct {
		Content any `json:"content"`
	}{Content: data}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, errs.InternalError(err, "error al serializar los datos para el servicio de PDF")
	}

	baseURL, err := s.resolveServiceURL(ctx)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, errs.InternalError(err, "error al construir URL del servicio de PDF")
	}
	u.Path = path.Join(u.Path, "/generatepdf/with_template")
	endpoint := u.String()

	var body bytes.Buffer
	bodyWriter := multipart.NewWriter(&body)

	if err := bodyWriter.WriteField("json", string(jsonData)); err != nil {
		return nil, errs.InternalError(err, "error al escribir campo json para el servicio de PDF")
	}

	fw, err := bodyWriter.CreateFormFile("template", "template.docx")
	if err != nil {
		return nil, errs.InternalError(err, "error al crear campo de archivo para el servicio de PDF")
	}

	if _, err := fw.Write(templateBytes); err != nil {
		return nil, errs.InternalError(err, "error al escribir archivo en body para el servicio de PDF")
	}

	if err := bodyWriter.Close(); err != nil {
		return nil, errs.InternalError(err, "error al cerrar body multipart del servicio de PDF")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return nil, errs.InternalError(err, "error al crear request al servicio de PDF")
	}
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errs.InternalError(err, "error al enviar request al servicio de PDF")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return io.ReadAll(resp.Body)
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		var response struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, errs.InternalError(err, "error al decodificar respuesta del servicio de PDF")
		}
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return nil, errs.BadRequestDirect(response.Message)
		case http.StatusInternalServerError:
			return nil, errs.InternalErrorDirect(response.Message)
		default:
			return nil, errs.InternalErrorDirect(response.Message)
		}
	}

	return nil, errs.InternalErrorDirect("error desconocido al procesar la respuesta del servicio de PDF")
}
