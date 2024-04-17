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

func SetIdentityServer(address string) { identityAddress = address }

func SetLogger(l *zap.Logger) { logger = l }

func ValidateToken(ctx context.Context, token string) (data *Data, err error) {
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
	var response IdentityServerResponse
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
const jwt_session_key = keyType("jwt-session-context-key")
const jwt_token_key = keyType("jwt-token-context-key")
const domain_key = keyType("domain_key")

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
		if data == nil {
			return answer.Err(c, errs.Bad("[close] session invalida"))
		}
		ctx := context.WithValue(c.Request().Context(), jwt_claims_key, data.Jwt)
		ctx = context.WithValue(ctx, jwt_session_key, data.Session)
		ctx = context.WithValue(ctx, jwt_token_key, token)
		ctx = context.WithValue(ctx, domain_key, data.Jwt.Empresa)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

const sucursal_codigo_key = keyType("sucursal_codigo_key")

// Este middleware verifica que el query param `sucursal` no esté vacío.
// Si el código de sucursal está vacío, devuelve un error de tipo `errs.Bad`.
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

func JwtClaims(c context.Context) (Jwt, bool) {
	values := c.Value(jwt_claims_key)
	if values == nil {
		return Jwt{}, false
	}
	v, ok := values.(Jwt)
	if !ok {
		if logger != nil {
			logger.Error("tokendata assert error")
		}
	}
	return v, ok
}

func ReadSession(c context.Context) (Session, bool) {
	values := c.Value(jwt_session_key)
	if values == nil {
		return Session{}, false
	}
	v, ok := values.(Session)
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
func TrabajadorAsociado(c context.Context) string {
	values, ok := JwtClaims(c)
	if !ok {
		return "####trabajador-sociado-no-found####"
	}
	return values.TabajadorCodigo
}

// Esta función toma una cadena y realiza una operación de split por el carácter '.'.
// Luego, devuelve el último segmento de la cadena resultante.
func RemovePrefix(s string) string {
	parts := strings.Split(strings.TrimSpace(s), ".")
	return strings.TrimSpace(parts[len(parts)-1:][0])
}
func CtxWithDomain(ctx context.Context, domain string) context.Context {
	return context.WithValue(ctx, domain_key, domain)
}

// Esta función concatena la cadena de la empresa con los sufijos proporcionados.
// Para una empresa "s1" y una lista de sufijos ["c1", "c2", "c3"], el resultado será "s1.c1.c2.c3".
func Empresa(c context.Context, suffix ...string) string {
	domain, ok := c.Value(domain_key).(string)
	if !ok {
		return "####empresa-no-found####"
	}
	var suff string
	for _, s := range suffix {
		suff += "." + RemovePrefix(s)
	}
	return domain + suff
}

func ReferenciaEmpresa(c context.Context) string {
	data, ok := JwtClaims(c)
	if !ok {
		return "####empresa-referencia-no-found####"
	}
	return data.ReferenciaEmpresa
}

func IntegracionURl(c context.Context) string {
	data, ok := JwtClaims(c)
	if !ok {
		return "####integracion-url-no-found####"
	}
	return data.IntegrationURL
}

func EmpresaPrefix(c context.Context) string  { return Empresa(c, "%") }
func SucursalPrefix(c context.Context) string { return Sucursal(c, "%") }

// Esta función concatena la empresa, la sucursal y los sufijos proporcionados en una cadena.
// Si la empresa es "e1", la sucursal es "s1" y los sufijos son ["c1", "c2", "c3"],
// el resultado será "e1.s1.c1.c2.c3".
func Sucursal(c context.Context, suffix ...string) string {
	value := c.Value(sucursal_codigo_key)
	sucursal, ok := value.(string)
	if !ok {
		return "####sucursal-no-found####"
	}
	return Empresa(c, append([]string{sucursal}, suffix...)...)
}

func Token(c context.Context) string {
	value := c.Value(jwt_token_key)
	token, ok := value.(string)
	if !ok {
		return "####token-undefined####"
	}
	return token
}

func CopyContext(ctx context.Context) context.Context {
	values, ok := JwtClaims(ctx)
	if !ok {
		return context.Background()
	}
	var newCtx = context.WithValue(context.Background(), jwt_claims_key, values)
	newCtx = context.WithValue(newCtx, domain_key, Empresa(ctx))
	newCtx = context.WithValue(newCtx, sucursal_codigo_key, Sucursal(ctx))
	return context.WithValue(newCtx, jwt_token_key, Token(ctx))
}
