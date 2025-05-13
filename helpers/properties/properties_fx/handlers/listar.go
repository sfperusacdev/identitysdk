package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/helpers/properties"
	"github.com/sfperusacdev/identitysdk/httpapi"
	"github.com/user0608/goones/answer"
)

type ListarPropertiesHandler struct {
	httpapi.MethodGet
	mutate properties.SystemPropertiesMutator
}

var _ httpapi.Route = (*ListarPropertiesHandler)(nil)

func NewListarPropertiesHandler(mutate properties.SystemPropertiesMutator) *ListarPropertiesHandler {
	return &ListarPropertiesHandler{mutate: mutate}
}

func (h *ListarPropertiesHandler) GetPath() string {
	return "/api/v1/_/system_properties"
}

func (h *ListarPropertiesHandler) HandleRequest(c echo.Context) error {
	records, err := h.mutate.RetriveAll(c.Request().Context())
	if err != nil {
		return answer.Err(c, err)
	}
	for i := range records {
		records[i].ID = identitysdk.RemovePrefix(records[i].ID)
	}
	return answer.Ok(c, records)
}
