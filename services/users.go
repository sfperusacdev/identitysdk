package services

import (
	"context"
	"strings"
	"time"

	"github.com/sfperusacdev/identitysdk/xreq"
)

type UserAccount struct {
	Code              string     `json:"code"`
	Username          string     `json:"username"`
	CompanyCode       string     `json:"company_code"`
	FullName          string     `json:"full_name"`
	IsDisabled        bool       `json:"is_disabled"`
	PasswordAttempts  int64      `json:"password_attempts"`
	ReferenceCode     string     `json:"reference_code"`
	ExternalReference string     `json:"external_reference"`
	LastLogin         *time.Time `json:"last_login"`
	CreatedAt         time.Time  `json:"created_at"`
	CreatedBy         string     `json:"created_by"`
	WriteAt           time.Time  `json:"write_at"`
	WriteBy           string     `json:"write_by"`
}

type UserAccountWithSubordinates struct {
	UserAccount
	Subordinates []UserAccount `json:"subordinates"`
}

func (s *ExternalBridgeService) GetUsuarios(ctx context.Context, domain string) ([]UserAccountWithSubordinates, error) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return []UserAccountWithSubordinates{}, nil
	}
	var response struct {
		Data []UserAccountWithSubordinates `json:"data"`
	}
	if err := xreq.MakeRequest(
		ctx,
		s.configProvider.Identity(),
		"/_/api/v1/internal/system/users",
		xreq.WithAccessToken(s.configProvider.IdentityAccessToken()),
		xreq.WithQueryParam("company_code", domain),
		xreq.WithUnmarshalResponseInto(&response),
	); err != nil {
		return nil, err
	}
	return response.Data, nil
}

type GrupoWithUsers struct {
	Code        string        `json:"code"`
	Description string        `json:"description"`
	Users       []UserAccount `json:"users"`
}

func (s *ExternalBridgeService) GetGrupos(ctx context.Context, domain string) ([]GrupoWithUsers, error) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return []GrupoWithUsers{}, nil
	}

	var response struct {
		Data []GrupoWithUsers `json:"data"`
	}

	if err := xreq.MakeRequest(
		ctx,
		s.configProvider.Identity(),
		"/_/api/v1/internal/system/groups",
		xreq.WithAccessToken(s.configProvider.IdentityAccessToken()),
		xreq.WithQueryParam("company_code", domain),
		xreq.WithUnmarshalResponseInto(&response),
	); err != nil {
		return nil, err
	}

	return response.Data, nil
}
