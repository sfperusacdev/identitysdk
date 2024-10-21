package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk"
)

type VerifyApiKey struct{}

func (*VerifyApiKey) Use() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{identitysdk.CheckApiKeyMiddleware}
}
