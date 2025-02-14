package httpapi

import (
	"context"

	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/permissions"
	"go.uber.org/fx"
)

var Module = fx.Module("http-server",
	fx.Provide(
		identitysdk.NewCheckJwtMiddleware,
		identitysdk.NewSucursalQueryParamMiddleware,
		identitysdk.NewCheckApiKeyMiddleware,
		permissions.NewPermissionMiddlewareBuilder,
		fx.Annotate(
			newEchoServer,
			fx.ParamTags(
				RouteTag,
			),
		),
	),
)

func newEchoServer(
	listRoutes []Route,
	jwtmiddle identitysdk.JwtMiddleware,
	sucursalMidl identitysdk.SucursalQueryParamMiddleware,
	permMiddlBld permissions.PermissionMiddlewareBuilder,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Middleware global
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=${time_unix}, method=${method}, uri=${uri}, status=${status}, ip=${remote_ip}, latency=${latency_human}\n",
	}))
	e.Use(middleware.Recover())

	for _, route := range listRoutes {
		middlewares := buildMiddlewares(route, jwtmiddle, sucursalMidl, permMiddlBld)
		routesMap := map[string]func(string, echo.HandlerFunc, ...echo.MiddlewareFunc) *echo.Route{
			http.MethodGet:    e.GET,
			http.MethodPost:   e.POST,
			http.MethodPut:    e.PUT,
			http.MethodDelete: e.DELETE,
			http.MethodPatch:  e.PATCH,
		}
		if handler, exists := routesMap[route.GetMethod()]; exists {
			handler(route.GetPath(), route.HandleRequest, middlewares...)
		} else {
			slog.Warn("Unsupported method found", "method", route.GetMethod())
		}
	}

	return e
}

func buildMiddlewares(
	route Route,
	jwtmiddle identitysdk.JwtMiddleware,
	sucursalMidl identitysdk.SucursalQueryParamMiddleware,
	permMiddlBld permissions.PermissionMiddlewareBuilder,
) []echo.MiddlewareFunc {
	middlewares := make([]echo.MiddlewareFunc, 0, 4)

	var isJwtProtected = false
	var ensureSessionf = func() {
		if !isJwtProtected {
			isJwtProtected = true
			middlewares = append(
				[]echo.MiddlewareFunc{echo.MiddlewareFunc(jwtmiddle)},
				middlewares...,
			)
		}
	}

	if _, ok := route.(publicRoute); !ok {
		ensureSessionf()
	}

	if _, ok := route.(sucursalValidator); !ok {
		ensureSessionf()
		middlewares = append(middlewares, echo.MiddlewareFunc(sucursalMidl))
	}

	if r, ok := route.(PermissionChecker); ok {
		if perms := r.CheckPermissions(); len(perms) > 0 {
			ensureSessionf()
			middlewares = append(middlewares, permMiddlBld(perms))
		}
	}

	if r, ok := route.(MiddlewaresProvider); ok {
		middlewares = append(middlewares, r.GetMiddlewares()...)
	}
	return middlewares
}

type ServeURLString string

func StartWebServer(lc fx.Lifecycle, e *echo.Echo, address ServeURLString) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				e.Use(middleware.CORS())
				slog.Info("Starting HTTP server", "address", address)

				if err := e.Start(string(address)); err != nil && err != http.ErrServerClosed {
					slog.Error("HTTP server error", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("Shutting down HTTP server")
			if err := e.Shutdown(ctx); err != nil {
				slog.Error("Error shutting down HTTP server", "error", err)
				return err
			}
			slog.Info("HTTP server stopped successfully")
			return nil
		},
	})
}
