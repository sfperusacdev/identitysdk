package facecropper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/sfperusacdev/identitysdk/services"
	"github.com/user0608/goones/errs"
)

type FaceCropService struct {
	bridge *services.ExternalBridgeService
}

func NewFaceCropService(bridge *services.ExternalBridgeService) *FaceCropService {
	return &FaceCropService{
		bridge: bridge,
	}
}

func (s *FaceCropService) resolveServiceURL(ctx context.Context) (string, error) {
	const variableName = "face_cropper_service_url"

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
			return "", errs.InternalErrorDirect("la variable `face_cropper_service_url` no fue encontrada")
		}
		return "", err
	}

	if validURL, ok := getValidURL(rawURL); ok {
		return validURL, nil
	}

	return "", errs.InternalErrorDirect("la variable `face_cropper_service_url` contiene un valor no válido")
}

func (s *FaceCropService) ProcessImage(ctx context.Context, imageBytes []byte) ([]byte, error) {
	baseURL, err := s.resolveServiceURL(ctx)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, errs.InternalError(err, errs.ErrInternal)
	}
	u.Path = path.Join(u.Path, "/facecrop")
	endpoint := u.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(imageBytes))
	if err != nil {
		return nil, errs.InternalError(err, errs.ErrInternal)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errs.InternalError(err, "error al conectar con facecrop")
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	if resp.StatusCode == http.StatusOK && strings.HasPrefix(contentType, "image/") {
		return io.ReadAll(resp.Body)
	}

	if strings.HasPrefix(contentType, "application/json") {
		var response struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, errs.InternalError(err, "error al leer respuesta de facecrop")
		}
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return nil, errs.BadRequestDirect(response.Message)
		case http.StatusInternalServerError:
			return nil, errs.InternalErrorDirect(response.Message)
		}
	}
	return nil, errs.InternalErrorDirect("error al procesar la imagen")
}
