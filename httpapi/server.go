package httpapi

import (
	"context"
	"fmt"

	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/permissions"
	"github.com/sfperusacdev/identitysdk/telemetry"
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
	if telemetryService := telemetry.Current(); telemetryService.Enabled() {
		e.Use(telemetryService.HTTPMiddleware())
	}
	e.Use(telemetry.AccessLogger)
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		DisableStackAll: true,
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			// Panic values may contain request secrets. Retain the failure
			// location without exporting the raw panic value.
			slog.ErrorContext(c.Request().Context(), "HTTP panic recovered",
				"error_type", fmt.Sprintf("%T", err), "stack", string(stack))
			return err
		},
	}))

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
	// Run after authentication and branch validation, never trusting raw headers
	// or query parameters to populate diagnostic tenant attributes.
	middlewares = append(middlewares, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			empresa, sucursal := identitysdk.Empresa_Sucursal(ctx)
			ctx = telemetry.WithTenantAttributes(ctx, empresa, sucursal)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	})

	return middlewares
}

type ServeURLString string

func StartWebServer(lc fx.Lifecycle, e *echo.Echo, address ServeURLString) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				e.Use(middleware.CORS())
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
