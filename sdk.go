package identitysdk

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/user0608/goones/answer"
	"github.com/user0608/goones/errs"
	"go.uber.org/zap"
)

var (
	identityAddress string
	logger          *zap.Logger
)

type identityServerResponse struct {
	Message string    `json:"message"`
	Data    tokenData `json:"data"`
}
type tokenData struct {
	Aud             []string `json:"aud"`
	Empresa         string   `json:"empresa"`
	Exp             int      `json:"exp"`
	Iat             int      `json:"iat"`
	ID              string   `json:"id"`
	Iss             string   `json:"iss"`
	TabajadorCodigo string   `json:"tabajador_codigo"`
	Username        string   `json:"username"`
	UsuarioReff     string   `json:"usuario_reff"`
	UsuarioCodigo   string   `json:"usuario_codigo"`
}

func SetIdentityServer(address string) { identityAddress = address }

func SetLogger(l *zap.Logger) { logger = l }

func ValidateToken(ctx context.Context, token string) (data *tokenData, err error) {
	hostUrl, err := url.JoinPath(identityAddress, "/v1/check-token")
	if err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		return nil, errs.Internal(errs.ErrInternal)
	}
	var buff bytes.Buffer
	var payload = struct {
		Token string `json:"token"`
	}{Token: token}
	if err := json.NewEncoder(&buff).Encode(&payload); err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		return nil, errs.Internal(errs.ErrInternal)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hostUrl, &buff)
	if err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		return nil, errs.Internal(errs.ErrInternal)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		return nil, errs.Internal("Auth server no responde")
	}
	defer res.Body.Close()
	var response identityServerResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		return nil, errs.Internal(errs.ErrInternal)
	}
	if res.StatusCode != http.StatusOK {
		return nil, errs.Bad(response.Message)
	}
	return &response.Data, nil
}

type keyType string

const jwt_claims_key = keyType("jwt-claims-context-key")

func CheckJwtMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			token = c.QueryParam("token")
		}
		if token == "" {
			return answer.Err(c, errs.Bad("[close] token no encontrado"))
		}
		data, err := ValidateToken(c.Request().Context(), token)
		if err != nil {
			return answer.Err(c, err)
		}
		ctx := context.WithValue(c.Request().Context(), jwt_claims_key, *data)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

const sucursal_codigo_key = keyType("sucursal_codigo_key")

// Este midleware verifica que hay un query param `sucursal no vacio`
func EnsureSucursalQueryParamMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		codigo := c.QueryParam("sucursal")
		if codigo == "" {
			return answer.Err(c, errs.Bad("el código de sucursal es necesario en los query params"))
		}
		ctx := context.WithValue(c.Request().Context(), sucursal_codigo_key, codigo)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

func JwtClaims(c context.Context) (tokenData, bool) {
	values := c.Value(jwt_claims_key)
	if values == nil {
		return tokenData{}, false
	}
	v, ok := values.(tokenData)
	if !ok {
		if logger != nil {
			logger.Error("tokendata assert error")
		}
	}
	return v, ok
}

func Username(c context.Context) string {
	values, ok := JwtClaims(c)
	if !ok {
		return "####username-no-found####"
	}
	return values.Username
}
func UsuarioReff(c context.Context) string {
	values, ok := JwtClaims(c)
	if !ok {
		return "####usuario-reff-no-found####"
	}
	return values.UsuarioReff
}

// Hace split por '.' y devuelve el último segmento
func RemovePrefix(s string) string {
	parts := strings.Split(strings.TrimSpace(s), ".")
	return strings.TrimSpace(parts[len(parts)-1:][0])
}

// Para una empresa s1 con sufix = [c1,c2,c3]
// el resultado sera s1.c1.c2.c3
func Empresa(c context.Context, suffix ...string) string {
	data, ok := JwtClaims(c)
	if !ok {
		return "####empresa-no-found####"
	}
	var suff string
	for _, s := range suffix {
		suff += "." + RemovePrefix(s)
	}
	return data.Empresa + suff
}
func EmpresaPrefix(c context.Context) string  { return Empresa(c, "%") }
func SucursalPrefix(c context.Context) string { return Sucursal(c, "%") }

// Si la empresa es : `e1` y la sucursal es `s1` y suffix = [c1,c2,c3]
// el resultado es e1.s1.c1.c3
func Sucursal(c context.Context, suffix ...string) string {
	value := c.Value(sucursal_codigo_key)
	sucursal, ok := value.(string)
	if !ok {
		return "####sucursal-no-found####"
	}
	return Empresa(c, append([]string{sucursal}, suffix...)...)
}

func CopyContext(ctx context.Context) context.Context {
	values, ok := JwtClaims(ctx)
	if !ok {
		return context.Background()
	}
	var newCtx = context.WithValue(context.Background(), jwt_claims_key, values)
	return context.WithValue(newCtx, sucursal_codigo_key, Sucursal(ctx))
}
