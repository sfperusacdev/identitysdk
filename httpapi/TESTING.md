# HTTP API Testing

`httpapi.NewTestServer` creates an `httptest.Server` using the same route
registration used by the application. It is intended for end-to-end tests of
projects that use this package.

The test server provides a fixed identity in every request context:

- Empresa: `empresa`
- Sucursal: `sucursal`
- Usuario: `usuario`
- Token: `token`

Authentication is simulated, so tests do not need a real Identity server.

## Basic Usage

Create the route using the normal `httpapi.Route` handlers and pass it to the
test server:

```go
package myapp_test

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/httpapi"
	"github.com/stretchr/testify/require"
)

func TestProfile(t *testing.T) {
	server := httpapi.NewTestServer(t, &httpapi.DefaultHandler{
		Method: http.MethodGet,
		Path:   "/profile",
		Handler: func(c echo.Context) error {
			ctx := c.Request().Context()
			empresa, sucursal := identitysdk.Empresa_Sucursal(ctx)

			return c.JSON(http.StatusOK, map[string]string{
				"empresa":  empresa,
				"sucursal": sucursal,
				"usuario":  identitysdk.Username(ctx),
				"token":    identitysdk.Token(ctx),
			})
		},
	})

	response, err := server.Client().Get(server.URL + "/profile")
	require.NoError(t, err)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)
}
```

The server is closed automatically through `testing.T.Cleanup`.

## Sucursal Routes

`DefaultSucursalHandler` receives the fixed test branch automatically. The
request does not need a `sucursal` query parameter:

```go
server := httpapi.NewTestServer(t, &httpapi.DefaultSucursalHandler{
	Method: http.MethodGet,
	Path:   "/branch",
	Handler: func(c echo.Context) error {
		_, sucursal := identitysdk.Empresa_Sucursal(c.Request().Context())
		return c.String(http.StatusOK, sucursal)
	},
})

response, err := server.Client().Get(server.URL + "/branch")
```

## Route Types

Use the same route types as the application:

- `DefaultHandler` for regular routes.
- `DefaultSucursalHandler` for routes that require a branch.
- `DefaultPublicHandler` for public routes.
- `DefaultApiKeyHandler` for API key protected routes.
- `DefaultAccessKeyHandler` for access key protected routes.
- `DefaultPublicClientHandler` for public client routes.

Route permissions are also evaluated using the test session. The fixed test
session has the `admin` permission, so permission-protected routes can be
tested without an external authorization service.

## Context Values

Inside a handler, the following values are available:

```go
ctx := c.Request().Context()

empresa, sucursal := identitysdk.Empresa_Sucursal(ctx)
usuario := identitysdk.Username(ctx)
token := identitysdk.Token(ctx)
```

The production HTTP server and its middleware are not changed by this helper.
