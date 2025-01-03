package binds

import (
	"github.com/labstack/echo/v4"
	"github.com/user0608/goones/errs"
)

func JSON(c echo.Context, payload any) error {
	if err := (&echo.DefaultBinder{}).BindBody(c, payload); err != nil {
		return errs.BadRequestf(err, "json document invalido")
	}
	return nil
}
func Query(c echo.Context, payload any) error {
	if err := (&echo.DefaultBinder{}).BindQueryParams(c, payload); err != nil {
		return errs.BadRequestf(err, "json document invalido")
	}
	return nil
}
