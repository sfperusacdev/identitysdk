package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk"
)

// Deprecated: en el futuro se eliminara
type VerifyJWT struct{}

// Deprecated: en el futuro se eliminara
func (*VerifyJWT) Use() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{echo.MiddlewareFunc(identitysdk.CheckJwtMiddleware)}
}

// Deprecated: en el futuro se eliminara
type VerifyJWTAndEnsureQueryParam struct {
}

// Deprecated: en el futuro se eliminara
func (*VerifyJWTAndEnsureQueryParam) Use() []echo.MiddlewareFunc {

	return []echo.MiddlewareFunc{echo.MiddlewareFunc(identitysdk.CheckJwtMiddleware), identitysdk.EnsureSucursalQueryParamMiddleware}
}

// Deprecated: en el futuro se eliminara
type SkipJwtMidd struct{}

// Deprecated: en el futuro se eliminara
func (*SkipJwtMidd) Use() []echo.MiddlewareFunc { return []echo.MiddlewareFunc{} }
