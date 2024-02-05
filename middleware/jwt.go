package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk"
)

type VerifyJWT struct{}

func (*VerifyJWT) Use() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{identitysdk.CheckJwtMiddleware}
}

type VerifyJWTAndEnsureQueryParam struct {
}

func (*VerifyJWTAndEnsureQueryParam) Use() []echo.MiddlewareFunc {

	return []echo.MiddlewareFunc{identitysdk.CheckJwtMiddleware, identitysdk.EnsureSucursalQueryParamMiddleware}
}

type SkipJwtMidd struct{}

func (*SkipJwtMidd) Use() []echo.MiddlewareFunc { return []echo.MiddlewareFunc{} }
