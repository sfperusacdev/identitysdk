package templates

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
	"github.com/sfperusacdev/identitysdk/telemetry"
	"github.com/sfperusacdev/identitysdk/xreq"
)

// TemplatesService communicates with the templates endpoints of the general service.
type TemplatesService struct {
	env bridgeidentity.IdentityProvider
}

func NewTemplatesService(env bridgeidentity.IdentityProvider) *TemplatesService {
	return &TemplatesService{env: env}
}

func (s *TemplatesService) endpoint(ctx context.Context, id uuid.UUID) (string, error) {
	baseURL, err := s.env.GetGeneralServiceURL(ctx, s.env.Empresa(ctx))
	if err != nil {
		return "", err
	}
	return url.JoinPath(baseURL, "/v1/api/templates", id.String())
}

// Get downloads the DOCX content of a template.
func (s *TemplatesService) Get(ctx context.Context, id uuid.UUID) ([]byte, error) {
	endpoint, err := s.endpoint(ctx, id)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", s.env.SessionToken(ctx))

	res, err := telemetry.DoHTTP(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("descargar template %s: HTTP %d", id, res.StatusCode)
	}
	return io.ReadAll(res.Body)
}

// Put uploads or replaces a template using the multipart "file" field.
func (s *TemplatesService) Put(ctx context.Context, id uuid.UUID, content []byte) error {
	endpoint, err := s.endpoint(ctx, id)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("file", id.String()+".docx")
	if err != nil {
		return err
	}
	if _, err := file.Write(content); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return xreq.MakeRequest(ctx, endpoint, "",
		xreq.WithMethod(http.MethodPut),
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
		xreq.WithHeader("Content-Type", writer.FormDataContentType()),
		xreq.WithRequestBody(&body),
	)
}

// Delete removes a template.
func (s *TemplatesService) Delete(ctx context.Context, id uuid.UUID) error {
	endpoint, err := s.endpoint(ctx, id)
	if err != nil {
		return err
	}
	return xreq.MakeRequest(ctx, endpoint, "",
		xreq.WithMethod(http.MethodDelete),
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
	)
}

// GetHash returns the SHA-256 hash stored for a template.
func (s *TemplatesService) GetHash(ctx context.Context, id uuid.UUID) (string, error) {
	endpoint, err := s.endpoint(ctx, id)
	if err != nil {
		return "", err
	}
	var response struct {
		Data string `json:"data"`
	}
	if err := xreq.MakeRequest(ctx, endpoint, "hash",
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
		xreq.WithUnmarshalResponseInto(&response),
	); err != nil {
		return "", err
	}
	return response.Data, nil
}
