package identitysdk

import (
	"context"
	"log/slog"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk/entities"
	"github.com/user0608/goones/answer"
	"github.com/user0608/goones/errs"
)

type keyType string

const jwt_claims_key = keyType("jwt-claims-context-key")
const jwt_claims_username = keyType("jwt_claims_username-key")
const jwt_session_key = keyType("jwt-session-context-key")
const jwt_token_key = keyType("jwt-token-context-key")
const domain_key = keyType("domain_key")

type JwtMiddleware echo.MiddlewareFunc

func CheckJwtMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			token = c.QueryParam("token")
		}
		if token == "" {
			return answer.Err(c, errs.BadRequestDirect("[close] token no encontrado"))
		}
		data, err := ValidateTokenWithCache(c.Request().Context(), token)
		if err != nil {
			return answer.Err(c, err)
		}
		if data == nil {
			return answer.Err(c, errs.BadRequestDirect("[close] session invalida"))
		}
		var newContext = BuildContext(c.Request().Context(), token, data)
		c.SetRequest(c.Request().WithContext(newContext))
		return next(c)
	}
}

func NewCheckJwtMiddleware() JwtMiddleware { return CheckJwtMiddleware }

func firstNoEmpty(vals ...string) string {
	for _, s := range vals {
		if s != "" {
			return s
		}
	}
	return ""
}

type ApiKeyMiddleware echo.MiddlewareFunc

func CheckApiKeyMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		apikey := firstNoEmpty(
			c.Request().Header.Get("x-api-key"),
			c.Request().Header.Get("X-API-KEY"),
		)
		if apikey == "" {
			return answer.Err(c, errs.BadRequestDirect("[close] API KEY no encontrado"))
		}
		data, err := ValidateApiKeyWithCache(c.Request().Context(), apikey)
		if err != nil {
			return answer.Err(c, err)
		}
		if data == nil {
			return answer.Err(c, errs.BadRequestDirect("[close] api key session invalida"))
		}
		var newContext = BuildApikeyContext(c.Request().Context(), apikey, &data.Apikey)
		c.SetRequest(c.Request().WithContext(newContext))
		return next(c)
	}
}

func NewCheckApiKeyMiddleware() ApiKeyMiddleware { return CheckApiKeyMiddleware }

// region ClientMiddleware
type JwtPublicClientMiddleware echo.MiddlewareFunc

func CheckJwtPublicClientMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			token = c.QueryParam("token")
		}
		if token == "" {
			return answer.Err(c, errs.BadRequestDirect("[close] token no encontrado"))
		}
		data, err := ValidatePublicClientToken(c.Request().Context(), token)
		if err != nil {
			return answer.Err(c, err)
		}
		if data == nil {
			return answer.Err(c, errs.BadRequestDirect("[close] session invalida"))
		}
		var newContext = BuildPublicClientContext(c.Request().Context(), token, data)
		c.SetRequest(c.Request().WithContext(newContext))
		return next(c)
	}
}

func NewCheckJwtPublicClientMiddleware() JwtPublicClientMiddleware {
	return CheckJwtPublicClientMiddleware
}

// region AccessKeyMiddleware
type AccessKeyMiddleware echo.MiddlewareFunc

func CheckAccessKeyMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		accesskey := c.Request().Header.Get("X-Access-Token")
		if accesskey == "" {
			return answer.Err(c, errs.BadRequestDirect("access-token no encontrado"))
		}
		if err := ValidateAccessKey(c.Request().Context(), accesskey); err != nil {
			return answer.Err(c, err)
		}
		return next(c)
	}
}

func NewCheckAccessKeyMiddleware() AccessKeyMiddleware {
	return CheckAccessKeyMiddleware
}

func BuildContext(ctx context.Context, token string, data *entities.JwtData) context.Context {
	newctx := context.WithValue(ctx, jwt_claims_key, data.Jwt)
	newctx = context.WithValue(newctx, jwt_claims_username, data.Jwt.Username)
	newctx = context.WithValue(newctx, jwt_session_key, data.Session)
	newctx = context.WithValue(newctx, jwt_token_key, token)
	newctx = context.WithValue(newctx, domain_key, data.Jwt.Empresa)
	return newctx
}

func BuildApikeyContext(ctx context.Context, apikey string, data *entities.Apikey) context.Context {
	newctx := context.WithValue(ctx, jwt_claims_key, entities.Jwt{Empresa: data.Empresa})
	newctx = context.WithValue(newctx, jwt_token_key, apikey)
	newctx = context.WithValue(newctx, domain_key, data.Empresa)
	return newctx
}

func BuildPublicClientContext(ctx context.Context, token string, data *entities.JwtPublicClientData) context.Context {
	newctx := context.WithValue(ctx, jwt_token_key, token)
	newctx = context.WithValue(newctx, jwt_claims_username, data.Jwt.Username)
	return newctx
}

const sucursal_codigo_key = keyType("sucursal_codigo_key")

// Este middleware verifica que el query param `sucursal` no esté vacío.
// Si el código de sucursal está vacío, devuelve un error de tipo `errs.Bad`.
type SucursalQueryParamMiddleware echo.MiddlewareFunc

func EnsureSucursalQueryParamMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		codigo := c.QueryParam("sucursal")
		if codigo == "" {
			return answer.Err(c, errs.BadRequestDirect("el código de sucursal es necesario en los query params"))
		}
		ctx := context.WithValue(c.Request().Context(), sucursal_codigo_key, codigo)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

func BuildContextWithSucursal(ctx context.Context, token, sucursal string, data *entities.JwtData) context.Context {
	ctx = BuildContext(ctx, token, data)
	return context.WithValue(ctx, sucursal_codigo_key, sucursal)
}

func NewSucursalQueryParamMiddleware() SucursalQueryParamMiddleware {
	return EnsureSucursalQueryParamMiddleware
}

func JwtClaims(c context.Context) (entities.Jwt, bool) {
	values := c.Value(jwt_claims_key)
	if values == nil {
		return entities.Jwt{}, false
	}
	v, ok := values.(entities.Jwt)
	if !ok {
		slog.Error("tokendata assert error")
	}
	return v, ok
}

func ReadSession(c context.Context) (entities.Session, bool) {
	values := c.Value(jwt_session_key)
	if values == nil {
		return entities.Session{}, false
	}
	v, ok := values.(entities.Session)
	if !ok {
		slog.Error("tokendata assert error")
	}
	return v, ok
}

func Username(c context.Context) string {
	username, ok := c.Value(jwt_claims_username).(string)
	if ok && username != "" {
		return username
	}
	return "####username-no-found####"
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
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ".")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func EqualIgnoringPrefix(s1, s2 string) bool {
	return RemovePrefix(s1) == RemovePrefix(s2)
}

func CtxWithDomain(ctx context.Context, domain string) context.Context {
	return context.WithValue(ctx, domain_key, domain)
}

func CtxWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, jwt_token_key, token)
}

// Esta función concatena la cadena de la empresa con los sufijos proporcionados.
// Para una empresa "s1" y una lista de sufijos ["c1", "c2", "c3"], el resultado será "s1.c1.c2.c3".
func Empresa(c context.Context, suffix ...string) string {
	domain, ok := c.Value(domain_key).(string)
	if !ok {
		domain = "####empresa-no-found####"
	}
	var suff string
	for _, s := range suffix {
		if s == "" {
			continue
		}
		suff += "." + RemovePrefix(s)
	}
	if suff == "" {
		return domain
	}
	return domain + suff
}

func EmpresaList(ctx context.Context, rawValues []string) []string {
	var normalized []string
	for _, v := range rawValues {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		normalized = append(normalized, Empresa(ctx, v))
	}
	return normalized
}

// Deprecated: ReferenciaEmpresa a sido deprecado, usar IntegracionExternaCodigo
// los nuevos metodos estan integrados en el servicio de ExternalBridgeService
func ReferenciaEmpresa(c context.Context) string {
	data, ok := JwtClaims(c)
	if !ok {
		return "####empresa-referencia-no-found####"
	}
	return data.ReferenciaEmpresa
}

// Deprecated: IntegracionURl a sido deprecado, usar IntegracionExternaURl
// los nuevos metodos estan integrados en el servicio de ExternalBridgeService
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
	var values = []string{}
	for _, v := range suffix {
		if v == "" {
			continue
		}
		values = append(values, v)
	}
	if len(values) == 0 {
		return ""
	}
	value := c.Value(sucursal_codigo_key)
	sucursal, ok := value.(string)
	if !ok {
		sucursal = "####sucursal-no-found####"
	}
	return Empresa(c, append([]string{sucursal}, values...)...)
}

func SucursalList(ctx context.Context, rawValues []string) []string {
	var normalized []string
	for _, v := range rawValues {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		normalized = append(normalized, Sucursal(ctx, v))
	}
	return normalized
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
