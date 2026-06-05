package httpapi

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type DefaultHandler struct {
	Method      string
	Path        string
	Permissions []string
	Handler     echo.HandlerFunc
}

var _ Route = (*DefaultHandler)(nil)
var _ PermissionChecker = (*DefaultHandler)(nil)

func (h *DefaultHandler) GetMethod() string {
	return defaultMethod(h.Method)
}

func (h *DefaultHandler) GetPath() string { return h.Path }

func (h *DefaultHandler) CheckPermissions() []string {
	return h.Permissions
}

func (h *DefaultHandler) HandleRequest(c echo.Context) error {
	return handleDefaultRequest(c, h.Handler)
}

type DefaultSucursalHandler struct {
	EnsureSucursal
	Method      string
	Path        string
	Permissions []string
	Handler     echo.HandlerFunc
}

var _ Route = (*DefaultSucursalHandler)(nil)
var _ PermissionChecker = (*DefaultSucursalHandler)(nil)

func (h *DefaultSucursalHandler) GetMethod() string {
	return defaultMethod(h.Method)
}

func (h *DefaultSucursalHandler) GetPath() string { return h.Path }

func (h *DefaultSucursalHandler) CheckPermissions() []string {
	return h.Permissions
}

func (h *DefaultSucursalHandler) HandleRequest(c echo.Context) error {
	return handleDefaultRequest(c, h.Handler)
}

type DefaultPublicHandler struct {
	PublicRoute
	Method      string
	Path        string
	Permissions []string
	Handler     echo.HandlerFunc
}

var _ Route = (*DefaultPublicHandler)(nil)
var _ PermissionChecker = (*DefaultPublicHandler)(nil)

func (h *DefaultPublicHandler) GetMethod() string {
	return defaultMethod(h.Method)
}

func (h *DefaultPublicHandler) GetPath() string { return h.Path }

func (h *DefaultPublicHandler) CheckPermissions() []string {
	return h.Permissions
}

func (h *DefaultPublicHandler) HandleRequest(c echo.Context) error {
	return handleDefaultRequest(c, h.Handler)
}

type DefaultApiKeyHandler struct {
	ApiKeyProtection
	Method      string
	Path        string
	Permissions []string
	Handler     echo.HandlerFunc
}

var _ Route = (*DefaultApiKeyHandler)(nil)
var _ PermissionChecker = (*DefaultApiKeyHandler)(nil)

func (h *DefaultApiKeyHandler) GetMethod() string {
	return defaultMethod(h.Method)
}

func (h *DefaultApiKeyHandler) GetPath() string { return h.Path }

func (h *DefaultApiKeyHandler) CheckPermissions() []string {
	return h.Permissions
}

func (h *DefaultApiKeyHandler) HandleRequest(c echo.Context) error {
	return handleDefaultRequest(c, h.Handler)
}

type DefaultAccessKeyHandler struct {
	AccessKeyProtection
	Method      string
	Path        string
	Permissions []string
	Handler     echo.HandlerFunc
}

var _ Route = (*DefaultAccessKeyHandler)(nil)
var _ PermissionChecker = (*DefaultAccessKeyHandler)(nil)

func (h *DefaultAccessKeyHandler) GetMethod() string {
	return defaultMethod(h.Method)
}

func (h *DefaultAccessKeyHandler) GetPath() string { return h.Path }

func (h *DefaultAccessKeyHandler) CheckPermissions() []string {
	return h.Permissions
}

func (h *DefaultAccessKeyHandler) HandleRequest(c echo.Context) error {
	return handleDefaultRequest(c, h.Handler)
}

type DefaultPublicClientHandler struct {
	PublicJwtClientProtection
	Method      string
	Path        string
	Permissions []string
	Handler     echo.HandlerFunc
}

var _ Route = (*DefaultPublicClientHandler)(nil)
var _ PermissionChecker = (*DefaultPublicClientHandler)(nil)

func (h *DefaultPublicClientHandler) GetMethod() string {
	return defaultMethod(h.Method)
}

func (h *DefaultPublicClientHandler) GetPath() string { return h.Path }

func (h *DefaultPublicClientHandler) CheckPermissions() []string {
	return h.Permissions
}

func (h *DefaultPublicClientHandler) HandleRequest(c echo.Context) error {
	return handleDefaultRequest(c, h.Handler)
}

func defaultMethod(method string) string {
	if method != "" {
		return method
	}
	return http.MethodGet
}

func handleDefaultRequest(c echo.Context, handler echo.HandlerFunc) error {
	if handler == nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"handler not configured",
		)
	}
	return handler(c)
}
