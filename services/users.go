package services

import (
	"context"
	"strings"

	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
	"github.com/sfperusacdev/identitysdk/xreq"
)

// Deprecated: use identity.UserAccount instead.
type UserAccount = bridgeidentity.UserAccount

// Deprecated: use identity.UserAccountWithSubordinates instead.
type UserAccountWithSubordinates = bridgeidentity.UserAccountWithSubordinates

// Deprecated: use bridge.Identity.GetUsuarios instead.
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

// Deprecated: use identity.GrupoWithUsers instead.
type GrupoWithUsers = bridgeidentity.GrupoWithUsers

// Deprecated: use bridge.Identity.GetGrupos instead.
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
