package httpapi

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/entities"
	"github.com/sfperusacdev/identitysdk/permissions"
)

const (
	testEmpresa  = "empresa"
	testSucursal = "sucursal"
	testUsuario  = "usuario"
	testToken    = "token"
)

// NewTestServer creates an HTTP server with the same route registration used
// by the application and a deterministic identity in every request context.
func NewTestServer(t *testing.T, routes ...Route) *httptest.Server {
	t.Helper()

	testAuth := identitysdk.JwtMiddleware(testAuthenticationMiddleware)
	e := newEchoServer(
		routes,
		testAuth,
		identitysdk.ApiKeyMiddleware(testAuthenticationMiddleware),
		identitysdk.JwtPublicClientMiddleware(testAuthenticationMiddleware),
		identitysdk.AccessKeyMiddleware(testAuthenticationMiddleware),
		identitysdk.SucursalQueryParamMiddleware(testSucursalMiddleware),
		permissions.NewPermissionMiddlewareBuilder(),
	)
	e.Use(testIdentityMiddleware)

	server := httptest.NewServer(e)
	t.Cleanup(server.Close)
	return server
}

func testAuthenticationMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}

func testSucursalMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := identitysdk.CtxWithSucursal(c.Request().Context(), testSucursal)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

func testIdentityMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := NewTestContext(c.Request().Context())

		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

// NewTestContext creates a context with the deterministic identity used by the
// test server, preserving values and cancellation from parent.
func NewTestContext(parent context.Context) context.Context {
	ctx := parent
	ctx = identitysdk.CtxWithJwtClaims(ctx, entities.Jwt{
		Empresa:  testEmpresa,
		Username: testUsuario,
		Zona:     "America/Lima",
	})
	ctx = identitysdk.CtxWithSession(ctx, entities.Session{
		Company:  testEmpresa,
		Username: testUsuario,
		Permissions: []entities.Permission{
			{ID: "admin", CompanyBrances: []string{testSucursal}},
		},
	})
	ctx = identitysdk.CtxWithDomain(ctx, testEmpresa)
	ctx = identitysdk.CtxWithUsername(ctx, testUsuario)
	ctx = identitysdk.CtxWithSucursal(ctx, testSucursal)
	ctx = identitysdk.CtxWithToken(ctx, testToken)
	return identitysdk.CtxWithRequestOrigin(ctx, "test")
}
