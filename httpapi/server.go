package httpapi

import (
	"context"
	"strings"

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
		identitysdk.NewCheckJwtPublicClientMiddleware,
		identitysdk.NewCheckAccessKeyMiddleware,
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
	apiKeymiddle identitysdk.ApiKeyMiddleware,
	jwtPublicClientMiddle identitysdk.JwtPublicClientMiddleware,
	accessKeyMiddleware identitysdk.AccessKeyMiddleware,

	sucursalMidl identitysdk.SucursalQueryParamMiddleware,
	permMiddlBld permissions.PermissionMiddlewareBuilder,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Middleware global
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=${time_unix}, method=${method}, uri=${uri}, status=${status}, ip=${remote_ip}, latency=${latency_human}\n",
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Path(), "/api/v1/_/system_properties")
		},
	}))
	e.Use(middleware.Recover())

	for _, route := range listRoutes {
		middlewares := buildMiddlewares(
			route,
			jwtmiddle,
			apiKeymiddle,
			jwtPublicClientMiddle,
			accessKeyMiddleware,
			sucursalMidl,
			permMiddlBld,
		)
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
	jwtMiddleware identitysdk.JwtMiddleware,
	apiKeyMiddleware identitysdk.ApiKeyMiddleware,
	publicJwtMiddleware identitysdk.JwtPublicClientMiddleware,
	accessKeyMiddleware identitysdk.AccessKeyMiddleware,

	sucursalMiddleware identitysdk.SucursalQueryParamMiddleware,
	permissionMiddlewareBuilder permissions.PermissionMiddlewareBuilder,
) []echo.MiddlewareFunc {
	middlewares := make([]echo.MiddlewareFunc, 0, 4)
	isRouteProtected := false

	if _, ok := route.(accessKeyProtect); ok {
		isRouteProtected = true
		middlewares = append(middlewares, echo.MiddlewareFunc(accessKeyMiddleware))
	}

	if _, ok := route.(publicClientJwtProtect); ok {
		isRouteProtected = true
		middlewares = append(middlewares, echo.MiddlewareFunc(publicJwtMiddleware))
	}

	if _, ok := route.(apikeyProtect); ok {
		isRouteProtected = true
		middlewares = append(middlewares, echo.MiddlewareFunc(apiKeyMiddleware))
	}

	if _, ok := route.(publicRoute); !ok {
		if !isRouteProtected {
			isRouteProtected = true
			middlewares = append(middlewares, echo.MiddlewareFunc(jwtMiddleware))
		}
	}

	if _, ok := route.(sucursalValidator); ok {
		if !isRouteProtected {
			isRouteProtected = true
			middlewares = append(middlewares, echo.MiddlewareFunc(jwtMiddleware))
		}
		middlewares = append(middlewares, echo.MiddlewareFunc(sucursalMiddleware))
	}

	if r, ok := route.(PermissionChecker); ok {
		if permissions := r.CheckPermissions(); len(permissions) > 0 {
			if !isRouteProtected {
				isRouteProtected = true
				middlewares = append(middlewares, echo.MiddlewareFunc(jwtMiddleware))
			}
			middlewares = append(middlewares, permissionMiddlewareBuilder(permissions))
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
