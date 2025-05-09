package httpapi

type accessKeyProtect interface {
	useaccessKeyProtection()
}

var _ accessKeyProtect = (*AccessKeyProtection)(nil)

type AccessKeyProtection struct{}

func (*AccessKeyProtection) useaccessKeyProtection() {}
