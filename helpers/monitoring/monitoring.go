package monitoring

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk/httpapi"
)

type HandlerNameHandler struct {
	httpapi.AccessKeyProtection
	httpapi.MethodGet
	service *MetricsService
}

var _ httpapi.Route = (*HandlerNameHandler)(nil)

func NewHandlerNameHandler(service *MetricsService) *HandlerNameHandler {
	return &HandlerNameHandler{service: service}
}

func (h *HandlerNameHandler) GetPath() string {
	return "/metrics"
}

func (h *HandlerNameHandler) HandleRequest(c echo.Context) error {
	resp := h.service.Collect()
	return c.JSON(http.StatusOK, resp)
}
