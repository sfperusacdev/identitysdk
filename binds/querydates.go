package binds

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk"
	"github.com/user0608/goones/errs"
	"github.com/user0608/goones/types"
)

var (
	dateFromQueryKeys = []string{"desde", "inicio", "from", "start", "left"}
	dateToQueryKeys   = []string{"hasta", "fin", "to", "end", "right"}
)

func QueryDateRangeUTC(c echo.Context) (time.Time, time.Time, error) {
	loc, err := requestLocation(c)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	from, err := requestDateOnlyAny(c, dateFromQueryKeys)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	to, err := requestDateOnlyAny(c, dateToQueryKeys)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return from.StartOfDayUTC(loc), to.EndOfDayUTC(loc), nil
}

var dayQueryKeys = []string{"fecha", "date", "dia", "day", "on", "at"}

func QuerySingleDateRangeUTC(c echo.Context) (time.Time, time.Time, error) {
	loc, err := requestLocation(c)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	date, err := requestDateOnlyAny(c, dayQueryKeys)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	start, end := date.ToUTCDayRange(loc)
	return start, end, nil
}

func requestLocation(c echo.Context) (*time.Location, error) {
	loc, err := identitysdk.Tz(c.Request().Context())
	if err != nil {
		return nil, errs.BadRequestError(err, "no se pudo obtener la zona horaria")
	}

	return loc, nil
}

func requestDateOnlyAny(c echo.Context, keys []string) (types.DateOnly, error) {
	for _, key := range keys {
		value := c.QueryParam(key)
		if value == "" {
			continue
		}

		date, err := types.NewDateOnlyFromString(value)
		if err != nil {
			return types.DateOnly{}, errs.BadRequestError(err, "fecha '%s' inválida", key)
		}

		return date, nil
	}

	return types.DateOnly{}, errs.BadRequestError(nil, "missing required param: expected one of %v", keys)
}
