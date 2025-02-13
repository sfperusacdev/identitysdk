package httpapi

type publicRoute interface {
	skipjwtmiddleware()
}

var _ publicRoute = (*PublicRoute)(nil)

type PublicRoute struct{}

func (*PublicRoute) skipjwtmiddleware() {}
