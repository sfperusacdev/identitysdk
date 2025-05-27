package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/xreq"
)

type Certificate struct {
	CommonName         string    `json:"common_name"`
	Country            string    `json:"country"`
	State              string    `json:"state"`
	Locality           string    `json:"locality"`
	Organization       string    `json:"organization"`
	OrganizationalUnit string    `json:"organizational_unit"`
	EmailAddress       string    `json:"email_address"`
	Issuer             string    `json:"issuer"`
	IssueDate          time.Time `json:"issue_date"`
	ExpirationDate     time.Time `json:"expiration_date"`
	SerialNumber       string    `json:"serial_number"`
	PublicKey          string    `json:"public_key"`
	Certificate        string    `json:"certificate"`
	PrivateKey         string    `json:"privatekey"`
}

type RequestPayload struct {
	CommonName         string `json:"common_name"`
	Country            string `json:"country"`
	State              string `json:"state"`
	Locality           string `json:"locality"`
	Organization       string `json:"organization"`
	OrganizationalUnit string `json:"organizational_unit"`
	EmailAddress       string `json:"email_address"`
}

func (s *ExternalBridgeService) IssueCertificate(ctx context.Context, payload RequestPayload) (*Certificate, error) {
	var domain = identitysdk.Empresa(ctx)
	return s.IssueCompanyCertificate(ctx, domain, payload)
}

// IssueCompanyCertificate Issues a new certificate for the given company
func (s *ExternalBridgeService) IssueCompanyCertificate(ctx context.Context, domain string, payload RequestPayload) (*Certificate, error) {
	var apiresponse struct {
		Message string      `json:"message"`
		Data    Certificate `json:"data"`
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		slog.Error("Error encoding JSON payload", "error", err)
		return nil, err
	}
	if err := xreq.MakeRequest(ctx,
		identitysdk.GetIdentityServer(),
		fmt.Sprintf("/api/v1/certificates/issue/%s", domain),
		xreq.WithMethod(http.MethodPost),
		xreq.WithAccessToken(identitysdk.GetAccessToken()),
		xreq.WithRequestBody(&buf),
		xreq.WithJsonContentType(),
		xreq.WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return nil, err
	}
	return &apiresponse.Data, nil
}

// GetCompanyCertificate retrieves the existing certificate for the authenticated company.
func (s *ExternalBridgeService) GetCompanyCertificate(ctx context.Context, domain string) (*Certificate, error) {
	var apiresponse struct {
		Message string      `json:"message"`
		Data    Certificate `json:"data"`
	}
	if err := xreq.MakeRequest(ctx,
		identitysdk.GetIdentityServer(),
		fmt.Sprintf("/api/v1/certificates/%s", domain),
		xreq.WithMethod(http.MethodGet),
		xreq.WithAccessToken(identitysdk.GetAccessToken()),
		xreq.WithJsonContentType(),
		xreq.WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return nil, err
	}
	return &apiresponse.Data, nil
}
