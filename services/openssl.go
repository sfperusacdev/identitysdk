package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sfperusacdev/identitysdk"
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

func (s *ExternalBridgeService) GenCertificate(ctx context.Context, cn string) (*Certificate, error) {
	if cn == "" {
		cn = "__"
	}
	var apiresponse struct {
		Message string      `json:"message"`
		Data    Certificate `json:"data"`
	}
	var token = identitysdk.Token(ctx)

	var err = s.makeRequest(ctx,
		identitysdk.GetIdentityServer(),
		fmt.Sprintf("/api/v1/certificates/%s", cn),
		token,
		&apiresponse,
	)
	if err != nil {
		return nil, err
	}
	return &apiresponse.Data, nil
}
