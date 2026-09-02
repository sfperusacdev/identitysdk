package identity

import (
	"context"

	"github.com/sfperusacdev/identitysdk"
)

type IdentityProvider struct{}

func NewIdentityProvider() IdentityProvider {
	return IdentityProvider{}
}

func (IdentityProvider) IdentityServer() string {
	return identitysdk.GetIdentityServer()
}

func (IdentityProvider) AccessToken() string {
	return identitysdk.GetAccessToken()
}

func (IdentityProvider) SessionToken(ctx context.Context) string {
	return identitysdk.Token(ctx)
}

func (IdentityProvider) Empresa(ctx context.Context, suffix ...string) string {
	return identitysdk.Empresa(ctx, suffix...)
}

func (IdentityProvider) Empresa_Sucursal(ctx context.Context) (string, string) {
	return identitysdk.Empresa_Sucursal(ctx)
}

func (IdentityProvider) Sucursal(ctx context.Context, suffix ...string) string {
	return identitysdk.Sucursal(ctx, suffix...)
}
