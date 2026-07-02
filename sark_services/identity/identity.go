package identity

import (
	"context"

	"github.com/sfperusacdev/identitysdk"
)

type IdentityProvider interface {
	IdentityServer() string
	AccessToken() string
	SessionToken(ctx context.Context) string
}

type DefaultIdentityProvider struct{}

func NewDefaultIdentityProvider() *DefaultIdentityProvider {
	return &DefaultIdentityProvider{}
}

func (DefaultIdentityProvider) IdentityServer() string {
	return identitysdk.GetIdentityServer()
}

func (DefaultIdentityProvider) AccessToken() string {
	return identitysdk.GetAccessToken()
}

func (DefaultIdentityProvider) SessionToken(ctx context.Context) string {
	return identitysdk.Token(ctx)
}
