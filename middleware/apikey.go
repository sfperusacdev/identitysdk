package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk"
)

// Deprecated: en el futuro se eliminara
type VerifyApiKey struct{}

// Deprecated: en el futuro se eliminara
func (*VerifyApiKey) Use() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{identitysdk.CheckApiKeyMiddleware}
}
