package httpapi

type sucursalValidator interface {
	ensure_sucursal_validation()
}

var _ sucursalValidator = (*EnsureSucursal)(nil)

type EnsureSucursal struct{}

func (*EnsureSucursal) ensure_sucursal_validation() {}
