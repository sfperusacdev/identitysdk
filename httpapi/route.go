package httpapi

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

const RouteTag = `group:"http-routes"`

type Route interface {
	GetMethod() string
	GetPath() string
	HandleRequest(c echo.Context) error
}

func AsRoute(fn any) any {
	return fx.Annotate(
		fn,
		fx.As(new(Route)),
		fx.ResultTags(RouteTag),
	)
}
