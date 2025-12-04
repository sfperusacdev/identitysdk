package querydates

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk"
	"github.com/user0608/goones/errs"
	"github.com/user0608/goones/types"
)

func ParseFechaRange(c echo.Context) (time.Time, time.Time, error) {
	desdeStr := c.QueryParam("desde")
	hastaStr := c.QueryParam("hasta")

	loc, err := identitysdk.Tz(c.Request().Context())
	if err != nil {
		return time.Time{}, time.Time{}, errs.BadRequestError(err, "no se pudo obtener la zona horaria")
	}

	desde, err := types.NewDateOnlyFromString(desdeStr)
	if err != nil {
		return time.Time{}, time.Time{}, errs.BadRequestError(err, "fecha 'desde' inválida")
	}

	hasta, err := types.NewDateOnlyFromString(hastaStr)
	if err != nil {
		return time.Time{}, time.Time{}, errs.BadRequestError(err, "fecha 'hasta' inválida")
	}

	return desde.StartOfDayUTC(loc), hasta.EndOfDayUTC(loc), nil
}

func ParseDiaRange(c echo.Context) (time.Time, time.Time, error) {
	fechaStr := c.QueryParam("fecha")

	loc, err := identitysdk.Tz(c.Request().Context())
	if err != nil {
		return time.Time{}, time.Time{}, errs.BadRequestError(err, "no se pudo obtener la zona horaria")
	}

	fecha, err := types.NewDateOnlyFromString(fechaStr)
	if err != nil {
		return time.Time{}, time.Time{}, errs.BadRequestError(err, "fecha inválida")
	}

	inicio, fin := fecha.ToUTCDayRange(loc)
	return inicio, fin, nil
}
