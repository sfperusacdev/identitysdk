package identity

import (
	"context"

	"github.com/sfperusacdev/identitysdk"
)

type IdentityProvider interface {
	IdentityServer() string
	AccessToken() string
	SessionToken(ctx context.Context) string
	Empresa(ctx context.Context, suffix ...string) string
	Empresa_Sucursal(ctx context.Context) (string, string)
	Sucursal(ctx context.Context, suffix ...string) string
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

func (DefaultIdentityProvider) Empresa(ctx context.Context, suffix ...string) string {
	return identitysdk.Empresa(ctx, suffix...)
}

func (DefaultIdentityProvider) Empresa_Sucursal(ctx context.Context) (string, string) {
	return identitysdk.Empresa_Sucursal(ctx)
}

func (DefaultIdentityProvider) Sucursal(ctx context.Context, suffix ...string) string {
	return identitysdk.Sucursal(ctx, suffix...)
}
