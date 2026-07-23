package httpapi

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk"
	"github.com/stretchr/testify/require"
)

func TestNewTestServerProvidesIdentityContext(t *testing.T) {
	server := NewTestServer(t, &DefaultHandler{
		Method: http.MethodGet,
		Path:   "/identity",
		Handler: func(c echo.Context) error {
			ctx := c.Request().Context()
			empresa, sucursal := identitysdk.Empresa_Sucursal(ctx)

			return c.JSON(http.StatusOK, map[string]string{
				"empresa":   empresa,
				"sucursal":  sucursal,
				"usuario":   identitysdk.Username(ctx),
				"token":     identitysdk.Token(ctx),
				"origin":    identitysdk.RequestOrigin(ctx),
				"test_flag": "true",
			})
		},
	})

	response, err := server.Client().Get(server.URL + "/identity")
	require.NoError(t, err)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)

	var body map[string]string
	require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	require.Equal(t, testEmpresa, body["empresa"])
	require.Equal(t, testSucursal, body["sucursal"])
	require.Equal(t, testUsuario, body["usuario"])
	require.Equal(t, testToken, body["token"])
	require.Equal(t, "test", body["origin"])
}

func TestNewTestServerUsesRouteMiddlewares(t *testing.T) {
	server := NewTestServer(t, &DefaultSucursalHandler{
		Method: http.MethodGet,
		Path:   "/branch",
		Handler: func(c echo.Context) error {
			_, sucursal := identitysdk.Empresa_Sucursal(c.Request().Context())
			return c.String(http.StatusOK, sucursal)
		},
	})

	response, err := server.Client().Get(server.URL + "/branch")
	require.NoError(t, err)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)
}
