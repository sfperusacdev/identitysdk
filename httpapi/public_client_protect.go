package httpapi

type publicClientJwtProtect interface {
	usepublicClientJwtProtection()
}

var _ publicClientJwtProtect = (*PublicJwtClientProtection)(nil)

type PublicJwtClientProtection struct{}

func (*PublicJwtClientProtection) usepublicClientJwtProtection() {}
