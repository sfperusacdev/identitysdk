# Crear endpoint

## Objetivo

Agregar un endpoint HTTP a un servicio Go siguiendo el estándar de la empresa.

Los servicios usan `identitysdk/setup`, por lo tanto el servidor HTTP ya está inicializado.  
Los endpoints se definen como rutas (`httpapi.Route`) y se registran mediante Fx usando `httpapi.AsRoute`.

---

## Ubicación de handlers

Para proyectos pequeños:

```text
internal/handlers/
```

Para proyectos medianos o grandes, usar contexto/boundary:

```text
internal/<context-name>/handlers/
```

Ejemplo:

```text
internal/almacen/handlers/
```

El package debe ser:

```go
package handlers
```

---

## Registro en Fx

Cada contexto debe tener un archivo:

```text
internal/<context-name>/module.go
```

Ejemplo:

```go
package almacen

import (
	"app/internal/almacen/domain"
	"app/internal/almacen/domain/repositories"
	"app/internal/almacen/domain/services"
	"app/internal/almacen/handlers"
	"app/internal/almacen/repos"
	"app/internal/almacen/servs"

	"github.com/sfperusacdev/identitysdk/httpapi"
	"go.uber.org/fx"
)

var Module = fx.Module("Almacen",
	fx.Provide(
		fx.Annotate(
			repos.NewProductoRepository,
			fx.As(new(repositories.ProductosRepository)),
		),
		fx.Annotate(
			servs.NewValidarProductoService,
			fx.As(new(services.ValidadorProductoService)),
		),
	),
	fx.Provide(
		domain.NewProductosUsecase,
	),
	fx.Provide(
		httpapi.AsRoute(handlers.CrearProductoHandler),
		httpapi.AsRoute(handlers.ListarProductosHandler),
	),
)
```

### Reglas para `module.go`

- Cada contexto debe exponer una variable `Module`.
- Usar `fx.Module("<NombreContexto>", ...)`.
- Registrar repositorios con `fx.Annotate` y `fx.As(...)`.
- Registrar servicios internos con `fx.Annotate` y `fx.As(...)`.
- Registrar usecases con `fx.Provide(...)`.
- Registrar handlers con:

```go
httpapi.AsRoute(handlers.NombreHandler)
```

- No registrar handlers manualmente en el servidor HTTP.
- No crear routers manuales.
- No crear `echo.Echo` manualmente.

---

## Librerías obligatorias en handlers

```go
"github.com/labstack/echo/v4"

"github.com/sfperusacdev/identitysdk"
"github.com/sfperusacdev/identitysdk/binds"
"github.com/sfperusacdev/identitysdk/httpapi"

"github.com/user0608/goones/answer"
"github.com/user0608/goones/errs"
"github.com/user0608/goones/kcheck"
```

---

## Patrón general de handler

```go
func NombreHandler(uc *domain.Usecase) httpapi.Route {
	return &httpapi.DefaultHandler{
		Method: http.MethodPost,
		Path:   "/resource",
		Handler: func(c echo.Context) error {
			ctx := c.Request().Context()

			// 1. parsear entrada
			// 2. validar
			// 3. aplicar multitenancy
			// 4. aplicar auditoría
			// 5. llamar usecase
			// 6. responder con answer
		},
	}
}
```

---

## Endpoint POST: crear

```go
func CrearPersonaHandler(uc *domain.PersonasUsecase) httpapi.Route {
	return &httpapi.DefaultHandler{
		Method: http.MethodPost,
		Path:   "/persona",
		Handler: func(c echo.Context) error {
			ctx := c.Request().Context()

			var persona models.PersonaModel
			if err := binds.JSON(c, &persona); err != nil {
				return answer.JsonErr(c)
			}

			if err := kcheck.Valid(persona); err != nil {
				return answer.Err(c, errs.BadRequestDirect(err.Error()))
			}

			persona.Codigo = identitysdk.Empresa(ctx, persona.Codigo)
			identitysdk.CreateBy(ctx, &persona)

			if err := uc.Crear(ctx, persona); err != nil {
				return answer.Err(c, err)
			}

			return answer.Success(c)
		},
	}
}
```

---

## Endpoint PUT/PATCH: actualizar

```go
func ActualizarPersonaHandler(uc *domain.PersonasUsecase) httpapi.Route {
	return &httpapi.DefaultHandler{
		Method: http.MethodPut,
		Path:   "/persona",
		Handler: func(c echo.Context) error {
			ctx := c.Request().Context()

			var persona models.PersonaModel
			if err := binds.JSON(c, &persona); err != nil {
				return answer.JsonErr(c)
			}

			if err := kcheck.Valid(persona); err != nil {
				return answer.Err(c, errs.BadRequestDirect(err.Error()))
			}

			persona.Codigo = identitysdk.Empresa(ctx, persona.Codigo)
			identitysdk.UpdateBy(ctx, &persona)

			if err := uc.Actualizar(ctx, persona); err != nil {
				return answer.Err(c, err)
			}

			return answer.Success(c)
		},
	}
}
```

---

## Endpoint GET: listado

Todo listado, búsqueda o consulta por empresa debe usar `identitysdk.EmpresaPrefix`.

```go
func ListarPersonaHandler(uc *domain.PersonasUsecase) httpapi.Route {
	return &httpapi.DefaultHandler{
		Method: http.MethodGet,
		Path:   "/persona",
		Handler: func(c echo.Context) error {
			ctx := c.Request().Context()

			prefix := identitysdk.EmpresaPrefix(ctx)

			personas, err := uc.Listar(ctx, prefix)
			if err != nil {
				return answer.Err(c, err)
			}

			return answer.Ok(c, personas)
		},
	}
}
```

---

## Endpoint DELETE: eliminación masiva

Por estándar, las eliminaciones deben soportar uno o muchos IDs/códigos.

El cliente no es confiable. Cualquier ID recibido desde body, query o path debe pasar por `identitysdk.Empresa`.

```go
func EliminarPersonaHandler(uc *domain.PersonasUsecase) httpapi.Route {
	return &httpapi.DefaultHandler{
		Method: http.MethodDelete,
		Path:   "/persona",
		Handler: func(c echo.Context) error {
			ctx := c.Request().Context()

			codigos, err := binds.RequestStrings(c)
			if err != nil {
				return answer.Err(c, err)
			}

			for i := range codigos {
				codigos[i] = identitysdk.Empresa(ctx, codigos[i])
			}

			if err := uc.Eliminar(ctx, codigos); err != nil {
				return answer.Err(c, err)
			}

			return answer.Success(c)
		},
	}
}
```

También puede usarse `binds.RequestUUIDs(c)` cuando los identificadores sean UUID.

---

## Endpoint con filtro por rango de fechas

El frontend envía fechas en la zona horaria de la empresa.  
El backend debe convertirlas a UTC usando los helpers de `binds`.

```go
func ListarPorFechaHandler(uc *domain.PersonasUsecase) httpapi.Route {
	return &httpapi.DefaultHandler{
		Method: http.MethodGet,
		Path:   "/persona/por-fecha",
		Handler: func(c echo.Context) error {
			ctx := c.Request().Context()

			from, to, err := binds.QueryDateRangeUTC(c)
			if err != nil {
				return answer.Err(c, err)
			}

			prefix := identitysdk.EmpresaPrefix(ctx)

			personas, err := uc.ListarPorFecha(ctx, prefix, from, to)
			if err != nil {
				return answer.Err(c, err)
			}

			return answer.Ok(c, personas)
		},
	}
}
```

Para una sola fecha:

```go
from, to, err := binds.QuerySingleDateRangeUTC(c)
```

---

## Reglas de multitenancy

Todos los sistemas son multitenant.

### Para crear o actualizar registros

Todo código/ID enviado por el cliente debe normalizarse con:

```go
codigo = identitysdk.Empresa(ctx, codigo)
```

Ejemplo:

```go
persona.Codigo = identitysdk.Empresa(ctx, persona.Codigo)
```

### Para uno o muchos IDs

```go
for i := range codigos {
	codigos[i] = identitysdk.Empresa(ctx, codigos[i])
}
```

### Para listar o consultar por empresa

Siempre usar:

```go
prefix := identitysdk.EmpresaPrefix(ctx)
```

Resultado esperado:

```text
sfperu.%
```

Este valor está preparado para consultas tipo:

```sql
WHERE codigo LIKE ?
```

### Para devolver códigos limpios al frontend

Cuando aplique, remover el prefijo interno:

```go
codigo = identitysdk.RemovePrefix(codigo)
```

---

## Reglas de contexto

Toda la información de sesión, empresa, usuario y zona horaria está en el contexto HTTP.

Siempre obtener contexto así:

```go
ctx := c.Request().Context()
```

Siempre propagarlo:

```text
handler → usecase → repository
```

Nunca usar:

```go
context.Background()
```

dentro de un flujo HTTP.

---

## Auditoría

Para creación:

```go
identitysdk.CreateBy(ctx, &entity)
```

Puebla campos como:

```text
created_at
created_by
```

Para actualización:

```go
identitysdk.UpdateBy(ctx, &entity)
```

Puebla campos como:

```text
updated_at
updated_by
```

---

## Datos de identidad disponibles

Todos se obtienen desde `ctx` y devuelven `string`:

```go
identitysdk.Username(ctx)
identitysdk.TrabajadorAsociado(ctx)
identitysdk.UsuarioReff(ctx)
```

### `UsuarioReff`

Usar `UsuarioReff` cuando se necesite el código del usuario representado en otro sistema, especialmente en integraciones.

---

## Entrada HTTP con `binds`

### Body JSON

```go
if err := binds.JSON(c, &payload); err != nil {
	return answer.JsonErr(c)
}
```

`binds.From(c, &payload)` también lee JSON body.

### Query params

```go
if err := binds.Query(c, &filter); err != nil {
	return answer.Err(c, err)
}
```

### IDs masivos string

```go
codigos, err := binds.RequestStrings(c)
```

Acepta campos como:

```json
{"codigo":"19239"}
```

```json
{"codigos":["1298","4455"]}
```

```json
{"ids":"abc123"}
```

### IDs masivos UUID

```go
ids, err := binds.RequestUUIDs(c)
```

### Rango de fechas

```go
from, to, err := binds.QueryDateRangeUTC(c)
```

Acepta:

```text
desde, inicio, from, start, left
hasta, fin, to, end, right
```

### Una sola fecha

```go
from, to, err := binds.QuerySingleDateRangeUTC(c)
```

Acepta:

```text
fecha, date, dia, day, on, at
```

---

## Validación

Validar payloads con:

```go
if err := kcheck.Valid(payload); err != nil {
	return answer.Err(c, errs.BadRequestDirect(err.Error()))
}
```

La definición de tags `chk` se documenta en:

```text
.ai/standards/validation.md
```

---

## Respuestas HTTP

Siempre usar `answer`.

### JSON inválido

```go
return answer.JsonErr(c)
```

### Error de validación

```go
return answer.Err(c, errs.BadRequestDirect(err.Error()))
```

### Error de usecase

```go
return answer.Err(c, err)
```

### Éxito sin payload

```go
return answer.Success(c)
```

### Éxito con payload

```go
return answer.Ok(c, data)
```

No usar directamente:

```go
c.JSON(...)
http.Error(...)
```

---

## Responsabilidades del handler

El handler solo puede:

1. Obtener `ctx`.
2. Parsear request con `binds`.
3. Validar con `kcheck`.
4. Aplicar multitenancy con `identitysdk`.
5. Aplicar auditoría con `identitysdk`.
6. Llamar al usecase.
7. Responder con `answer`.

El handler no debe contener lógica de negocio.

---

## Anti-patterns

No crear servidor HTTP:

```go
http.ListenAndServe(":8080", handler)
```

No responder manualmente:

```go
return c.JSON(http.StatusOK, data)
```

No usar contexto global:

```go
ctx := context.Background()
```

No acceder a DB desde handler:

```go
db.Query(...)
```

No confiar en IDs enviados por cliente:

```go
if err := uc.Eliminar(ctx, codigos); err != nil {
	return answer.Err(c, err)
}
```

No listar sin prefix:

```go
personas, err := uc.Listar(ctx)
```

No parsear fechas manualmente:

```go
time.Parse("2006-01-02", c.QueryParam("desde"))
```

No registrar rutas manualmente:

```go
e.POST("/persona", handler)
```

---

## Checklist

- [ ] Handler ubicado en `internal/handlers` o `internal/<context-name>/handlers`.
- [ ] Función retorna `httpapi.Route`.
- [ ] Usa `httpapi.DefaultHandler`.
- [ ] Handler registrado en `module.go` con `httpapi.AsRoute`.
- [ ] Usa `ctx := c.Request().Context()`.
- [ ] Propaga `ctx` al usecase.
- [ ] Usa `binds` para entrada.
- [ ] Usa `kcheck.Valid` para validación cuando aplique.
- [ ] Usa `identitysdk.Empresa` para todo ID/código recibido del cliente.
- [ ] Usa `identitysdk.EmpresaPrefix` para todo listado/consulta por empresa.
- [ ] Usa `identitysdk.CreateBy` en creación.
- [ ] Usa `identitysdk.UpdateBy` en actualización.
- [ ] Usa `answer` para todas las respuestas.
- [ ] No usa `c.JSON`.
- [ ] No usa `http.ListenAndServe`.
- [ ] No usa `context.Background`.
- [ ] No accede directamente a DB.
- [ ] No registra rutas manualmente en Echo.

---

## Prompt recomendado para IA

```text
Crea un endpoint siguiendo estrictamente `.ai/recipes/create-endpoint.md`.

Usa `httpapi.Route` y `httpapi.DefaultHandler`.
Registra el handler en `module.go` usando `httpapi.AsRoute`.
Usa `binds` para parseo.
Usa `kcheck.Valid` para validación.
Usa `identitysdk` para multitenancy, auditoría e identidad.
Todo ID/código recibido del cliente debe pasar por `identitysdk.Empresa`.
Todo listado debe usar `identitysdk.EmpresaPrefix`.
Usa `answer` para respuestas.
No uses `c.JSON`, `http.ListenAndServe`, `context.Background` ni rutas Echo manuales.
No pongas lógica de negocio en el handler.
```