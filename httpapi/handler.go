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

func NewHandler() *DefaultHandler {
	return &DefaultHandler{}
}

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
