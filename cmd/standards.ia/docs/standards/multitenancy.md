# Multitenancy

## Objetivo

Estandarizar cómo los servicios manejan datos por empresa (tenant).

Todos los sistemas son multitenant.  
La empresa actual se obtiene desde el contexto del request HTTP.

```go
ctx := c.Request().Context()
```

---

## Regla base

Toda la información de:

- empresa
- usuario
- sesión
- zona horaria

vive en el `context.Context`.

El contexto debe propagarse siempre:

```text
handler → usecase → repository
```

Nunca usar:

```go
context.Background()
```

dentro de un flujo HTTP.

---

## Código multitenant

Los códigos internos se almacenan con prefijo de empresa.

Ejemplo:

```text
empresa: sfperu
codigo recibido: 73491346
codigo interno: sfperu.73491346
```

---

## Generación de códigos

Todos los identificadores (códigos) deben ser proporcionados por el cliente (frontend).

El backend no debe generar códigos automáticamente.

---

## Reglas obligatorias

- No usar campos autoincrementales como identificadores de negocio.
- No generar códigos en handler.
- No generar códigos en usecase.
- No generar códigos en repository.
- El código siempre debe venir en el payload del cliente.

---

## Ejemplo correcto

```go
var payload models.PersonaModel

if err := binds.JSON(c, &payload); err != nil {
	return answer.JsonErr(c)
}

payload.Codigo = identitysdk.Empresa(ctx, payload.Codigo)
```

---

## Ejemplo incorrecto

No generar códigos en backend:

```go
persona.Codigo = uuid.New().String()
```

```go
persona.Codigo = generarCodigo()
```

No depender de autoincrementales:

```sql
id SERIAL PRIMARY KEY
```

---

## Relación con multitenancy

Flujo correcto:

```text
frontend → envía codigo → backend aplica Empresa → se persiste
```

Ejemplo:

```text
frontend: 73491346
backend:  sfperu.73491346
```

---

## Validación de códigos

El código debe validarse como cualquier otro campo:

```go
Codigo string `chk:"required"`
```

---

## `identitysdk.Empresa`

Usar para construir identificadores multitenant.

```go
codigo := identitysdk.Empresa(ctx, codigo)
```

Ejemplo:

```go
persona.Codigo = identitysdk.Empresa(ctx, persona.Codigo)
```

Soporta múltiples sufijos:

```go
codigo := identitysdk.Empresa(ctx, "c1", "c2")
```

Resultado:

```text
sfperu.c1.c2
```

---

## IDs recibidos del cliente

Todo ID o código recibido desde:

- body
- query
- path

debe pasar por:

```go
identitysdk.Empresa(ctx, value)
```

---

## Ejemplo con múltiples IDs

```go
for i := range codigos {
	codigos[i] = identitysdk.Empresa(ctx, codigos[i])
}
```

---

## `identitysdk.EmpresaPrefix`

Usar para listados, búsquedas y consultas por empresa.

```go
prefix := identitysdk.EmpresaPrefix(ctx)
```

Resultado:

```text
sfperu.%
```

Uso en repository:

```sql
WHERE codigo LIKE ?
```

```go
rs := tx.Raw(qry, prefix).Scan(&items)
```

---

## Regla crítica para prefix

`identitysdk.EmpresaPrefix(ctx)` ya devuelve el formato correcto.

No agregar `%` manualmente.

Correcto:

```go
prefix := identitysdk.EmpresaPrefix(ctx)
```

Incorrecto:

```go
prefix := identitysdk.EmpresaPrefix(ctx) + "%"
```

---

## `identitysdk.RemovePrefix`

Usar cuando se necesite devolver datos sin el prefijo interno.

```go
codigo := identitysdk.RemovePrefix("sfperu.23932")
```

Resultado:

```text
23932
```

---

## Crear registros

```go
payload.Codigo = identitysdk.Empresa(ctx, payload.Codigo)
identitysdk.CreateBy(ctx, &payload)

if err := uc.CrearPersona(ctx, payload); err != nil {
	return answer.Err(c, err)
}
```

---

## Actualizar registros

```go
payload.Codigo = identitysdk.Empresa(ctx, payload.Codigo)
identitysdk.UpdateBy(ctx, &payload)

if err := uc.ActualizarPersona(ctx, payload); err != nil {
	return answer.Err(c, err)
}
```

---

## Listar registros

Todo listado debe usar `EmpresaPrefix`.

```go
prefix := identitysdk.EmpresaPrefix(ctx)

data, err := uc.ListarPersonas(ctx, prefix)
if err != nil {
	return answer.Err(c, err)
}
```

---

## Eliminar registros

```go
codigos, err := binds.RequestStrings(c)
if err != nil {
	return answer.Err(c, err)
}

for i := range codigos {
	codigos[i] = identitysdk.Empresa(ctx, codigos[i])
}

if err := uc.EliminarPersonas(ctx, codigos); err != nil {
	return answer.Err(c, err)
}
```

---

## Zona horaria

Cada empresa puede tener su propia zona horaria.

```go
identitysdk.Tz(ctx)
```

Los helpers de `binds` convierten fechas a UTC usando esa zona:

```go
binds.QueryDateRangeUTC(c)
binds.QuerySingleDateRangeUTC(c)
```

---

## Datos de identidad

Disponibles desde `context.Context`:

```go
identitysdk.Username(ctx)
identitysdk.TrabajadorAsociado(ctx)
identitysdk.UsuarioReff(ctx)
```

---

## Responsabilidad por capa

### Handler

Debe:

- obtener `ctx`
- transformar IDs con `identitysdk.Empresa`
- usar `EmpresaPrefix` en listados
- aplicar `CreateBy` / `UpdateBy`
- propagar `ctx`

---

### Usecase

Debe:

- recibir `ctx`
- recibir IDs ya transformados
- recibir `prefix` cuando aplique
- no calcular tenant

---

### Repository

Debe:

- recibir `ctx`
- usar `prefix` en queries
- usar IDs ya transformados

---

## Anti-patterns

No usar contexto global:

```go
ctx := context.Background()
```

No generar códigos:

```go
codigo := uuid.New().String()
```

No usar autoincrementales:

```sql
id SERIAL PRIMARY KEY
```

No agregar `%` manualmente:

```go
prefix := identitysdk.EmpresaPrefix(ctx) + "%"
```

No listar sin prefix:

```go
uc.Listar(ctx)
```

No enviar IDs sin transformar:

```go
uc.Eliminar(ctx, codigos)
```

---

## Checklist

- [ ] Se usa `ctx := c.Request().Context()`
- [ ] Se propaga el contexto
- [ ] Los códigos vienen del frontend
- [ ] No se generan códigos en backend salvo sea necesario o se parte del caso de uso
- [ ] Los IDs pasan por `identitysdk.Empresa`
- [ ] Los listados usan `EmpresaPrefix`
- [ ] No se agrega `%` manualmente
- [ ] Se usa `CreateBy` / `UpdateBy`
- [ ] Se usa `RemovePrefix` cuando aplica
- [ ] Se usa `Tz` para fechas cuando aplica