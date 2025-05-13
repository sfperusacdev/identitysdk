package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk/binds"
	"github.com/sfperusacdev/identitysdk/helpers/properties"
	"github.com/sfperusacdev/identitysdk/helpers/properties/models"
	"github.com/sfperusacdev/identitysdk/httpapi"
	"github.com/user0608/goones/answer"
)

type UpdatePropertiesHandler struct {
	httpapi.MethodPut
	mutate properties.SystemPropertiesMutator
}

var _ httpapi.Route = (*UpdatePropertiesHandler)(nil)

func NewUpdatePropertiesHandler(mutate properties.SystemPropertiesMutator) *UpdatePropertiesHandler {
	return &UpdatePropertiesHandler{mutate: mutate}
}

func (h *UpdatePropertiesHandler) GetPath() string {
	return "/api/v1/_/system_properties"
}

func (h *UpdatePropertiesHandler) HandleRequest(c echo.Context) error {
	var req struct {
		Items []models.BasicSystemProperty `json:"items"`
	}
	if err := binds.JSON(c, &req); err != nil {
		return answer.JsonErr(c)
	}
	if err := h.mutate.Update(c.Request().Context(), req.Items); err != nil {
		return answer.Err(c, err)
	}
	return answer.Success(c)
}
