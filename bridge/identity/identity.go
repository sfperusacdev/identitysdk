package identity

import (
	"context"

	"github.com/sfperusacdev/identitysdk"
)

type Provider interface {
	IdentityServer() string
	AccessToken() string
	SessionToken(ctx context.Context) string
}

type DefaultProvider struct{}

func NewDefaultProvider() *DefaultProvider {
	return &DefaultProvider{}
}

func (DefaultProvider) IdentityServer() string {
	return identitysdk.GetIdentityServer()
}

func (DefaultProvider) AccessToken() string {
	return identitysdk.GetAccessToken()
}

func (DefaultProvider) SessionToken(ctx context.Context) string {
	return identitysdk.Token(ctx)
}
