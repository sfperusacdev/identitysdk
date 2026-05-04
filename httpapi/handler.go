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
	if h.Method != "" {
		return h.Method
	}
	return http.MethodGet
}

func (h *DefaultHandler) GetPath() string { return h.Path }

func (h *DefaultHandler) CheckPermissions() []string {
	return h.Permissions
}

func (h *DefaultHandler) HandleRequest(c echo.Context) error {
	if h.Handler == nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"handler not configured",
		)
	}
	return h.Handler(c)
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
	if h.Method != "" {
		return h.Method
	}
	return http.MethodGet
}

func (h *DefaultSucursalHandler) GetPath() string { return h.Path }

func (h *DefaultSucursalHandler) CheckPermissions() []string {
	return h.Permissions
}

func (h *DefaultSucursalHandler) HandleRequest(c echo.Context) error {
	if h.Handler == nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"handler not configured",
		)
	}
	return h.Handler(c)
}
