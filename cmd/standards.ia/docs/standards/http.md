# HTTP Standards

## Objetivo

Estandarizar el uso de HTTP en los servicios.

---

## Framework HTTP

El framework HTTP estándar es Echo.

```go
"github.com/labstack/echo/v4"
```

El servidor es inicializado por:

```go
identitysdk/setup
```

No se debe crear `echo.New()` manualmente.

---

## Versionado de API

Todas las rutas deben estar versionadas.

Formato:

```text
/api/v1/<recurso>
```

Ejemplo:

```text
/api/v1/personas
/api/v1/personas/:codigo
```

---

## Reglas de rutas

- Usar sustantivos en plural
- No usar verbos
- Usar minúsculas
- Seguir REST
- Siempre incluir `/api/v1`

---

## Path params vs Query params

### Path params (Echo)

Echo usa `:param`

```text
/api/v1/personas/:codigo
```

Uso:

```go
codigo := c.Param("codigo")
```

Casos:

- obtener por ID
- eliminar por ID
- operaciones directas sobre recurso

---

### Query params

Para filtros, búsqueda, rangos, paginación:

```text
/api/v1/personas?estado=activo
/api/v1/personas?desde=2024-01-01&hasta=2024-01-31
```

Uso:

```go
c.QueryParam("estado")
```

---

## Regla de decisión

```text
identificador único → path param
filtros/búsqueda    → query param
```

---

## Métodos HTTP

- GET     → obtener datos
- POST    → crear
- PUT     → actualizar completo
- PATCH   → parcial
- DELETE  → eliminar

---

## Ejemplos

```http
POST   /api/v1/personas
GET    /api/v1/personas
GET    /api/v1/personas/:codigo
PUT    /api/v1/personas
DELETE /api/v1/personas
```

---

## Uso de Echo

Handlers:

```go
func(c echo.Context) error
```

Obtener contexto:

```go
ctx := c.Request().Context()
```

Obtener path param:

```go
codigo := c.Param("codigo")
```

Obtener query param:

```go
estado := c.QueryParam("estado")
```

---

## Registro de rutas

Se hace mediante:

```go
httpapi.Route
httpapi.DefaultHandler
```

En `module.go`:

```go
httpapi.AsRoute(handlers.CrearPersonaHandler)
```

No usar `e.GET`, `e.POST`.

---

## Respuestas

Usar siempre:

```go
return answer.Ok(c, data)
return answer.Success(c)
return answer.Err(c, err)
```

---

## Status codes

Manejados por `errs` y `answer`.

---

## Reglas obligatorias

- Todas las rutas usan `/api/v1`
- Usar plural (`personas`)
- Path params con `:param`
- No usar `{param}`
- Query params para filtros
- No usar `c.JSON`
- No usar `http.Error`
- No crear servidor manual
- No registrar rutas manualmente
- No pasar `echo.Context` al usecase

---

## Anti-patterns

No usar formato incorrecto de params:

```text
/api/v1/personas/{codigo}
```

No usar query para ID:

```text
/api/v1/personas?id=123
```

No crear Echo manual:

```go
e := echo.New()
```

No registrar rutas manualmente:

```go
e.GET("/api/v1/personas", handler)
```

No responder manualmente:

```go
return c.JSON(200, data)
```

No usar echo fuera del handler:

```go
func (uc *Usecase) Crear(c echo.Context) error
```

---

## Checklist

- [ ] Usa `/api/v1`
- [ ] Usa plural
- [ ] Usa `:param` (Echo)
- [ ] No usa `{param}`
- [ ] Usa query params correctamente
- [ ] Usa `answer`
- [ ] No usa `c.JSON`
- [ ] No usa `http.Error`