package httpapi

import "github.com/labstack/echo/v4"

type MiddlewaresProvider interface {
	GetMiddlewares() []echo.MiddlewareFunc
}
