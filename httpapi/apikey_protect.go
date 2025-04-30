package httpapi

type apikeyProtect interface {
	useapikeyProtection()
}

var _ apikeyProtect = (*ApiKeyProtection)(nil)

type ApiKeyProtection struct{}

func (*ApiKeyProtection) useapikeyProtection() {}
