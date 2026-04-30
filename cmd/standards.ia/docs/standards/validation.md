# Validación con kcheck

## Objetivo

Estandarizar la validación de structs usando tags `chk`.

La librería estándar para validación es:

```go
"github.com/user0608/goones/kcheck"
```

---

## Uso de kcheck

El uso de `kcheck` no es obligatorio en todos los endpoints.

Debe utilizarse cuando sea necesario validar:

- campos requeridos
- formato (email, uuid, url, etc.)
- longitud o rangos
- reglas simples de estructura

---

## Cuándo usar kcheck

Usar `kcheck` cuando:

- el payload tiene validaciones declarativas claras
- se requiere consistencia en validación
- se desea evitar validación manual repetitiva

Ejemplo:

```go
if err := kcheck.Valid(payload); err != nil {
	return answer.Err(c, errs.BadRequestDirect(err.Error()))
}
```

---

## Cuándo NO usar kcheck

No es necesario usar `kcheck` cuando:

- el payload es simple y no requiere validación
- la validación depende de lógica de negocio
- la validación requiere acceso a DB o servicios externos

En esos casos, validar en el usecase.

---

## Regla de decisión

```text
validación estructural → handler (kcheck opcional)
validación de negocio → usecase
```

## Uso en handlers

Todo payload recibido por HTTP debe validarse después de parsearse con `binds`.

```go
var payload models.PersonaModel
if err := binds.JSON(c, &payload); err != nil {
	return answer.JsonErr(c)
}

if err := kcheck.Valid(payload); err != nil {
	return answer.Err(c, errs.BadRequestDirect(err.Error()))
}
```

---

## Tags básicos

### Requeridos

```go
Codigo string `chk:"required"`
Nombre string `chk:"nonil"`
```

Tags disponibles:

```text
required
nonil
```

---

## Longitud y tamaño

```go
Codigo string `chk:"len=8"`
Nombre string `chk:"min=2 max=80"`
```

Tags disponibles:

```text
len=n
min=n
max=n
```

---

## Comparadores numéricos

```go
Edad int `chk:"gte=18 lte=120"`
```

Tags disponibles:

```text
gt=n
gte=n
lt=n
lte=n
```

---

## Strings

```go
Codigo string `chk:"upper alphanum len=6"`
```

Tags disponibles:

```text
alpha
alphanum
num
decimal
lower
upper
```

---

## Formatos

```go
Email string `chk:"required email"`
ID    string `chk:"uuid"`
URL   string `chk:"url"`
```

Tags disponibles:

```text
email
uuid
url
ip
ipv4
ipv6
```

---

## Opciones permitidas

```go
Estado string `chk:"required oneof=active,inactive,pending"`
```

---

## Prefijos, sufijos y contenido

```go
Codigo string `chk:"prefix=PER"`
Nombre string `chk:"contains=SA"`
Clave  string `chk:"suffix=001"`
```

---

## Fechas

```go
Fecha      string    `chk:"date"`
Hora       string    `chk:"time"`
FechaHora  string    `chk:"datetime"`
CreatedAt  time.Time `chk:"utc"`
```

Formatos:

```text
date      → 2006-01-02
time      → 15:04:05
datetime  → 2006-01-02 15:04:05
utc       → 2026-04-30T15:04:05Z
```

---

## Tipos soportados

```text
string
*string
int
uint
float
bool
time.Time
punteros
structs anidados
```

---

## Validación parcial

Para omitir campos:

```go
err := kcheck.Valid(payload, "Email")
```

Para validar solo campos específicos:

```go
err := kcheck.ValidSelect(payload, "Nombre")
```

---

## Structs anidados

```go
type Address struct {
	City string `chk:"required min=3"`
}

type User struct {
	Name    string  `chk:"required min=2 max=50"`
	Email   string  `chk:"required email"`
	Address Address
}
```

---

## Custom validators

Solo crear validadores personalizados cuando no exista un tag estándar.

```go
v := kcheck.New()

v.Register("startsx", func(f kcheck.Field) error {
	if s, ok := f.Value.(string); ok {
		if !strings.HasPrefix(s, "x") {
			return fmt.Errorf("debe empezar con x")
		}
	}
	return nil
})
```

Uso:

```go
type DTO struct {
	Code string `chk:"startsx"`
}

err := v.Struct(DTO{Code: "abc"})
```

---

## Reglas obligatorias

- Validar payloads HTTP en el handler.
- Validar después de `binds.JSON` o `binds.Query`.
- Responder errores de validación con:

```go
return answer.Err(c, errs.BadRequestDirect(err.Error()))
```

- Usar tags `chk` en `models` o `dtos`.
- No validar manualmente en handler si existe un tag `chk`.
- No usar validadores externos si `kcheck` cubre el caso.
- Usar `dtos` solo cuando el modelo no represente bien la entrada o salida.

---

## Usecase vs Handler

### Handler

Valida estructura del request:

```text
campos requeridos
formato
longitud
tipo
rango básico
```

### Usecase

Valida reglas de negocio:

```text
existencia en DB
permisos
estado permitido
reglas cruzadas
integraciones externas
```

---

## Anti-patterns

No validar manualmente si existe tag:

```go
if payload.Email == "" {
	return answer.Err(c, errs.BadRequestDirect("email requerido"))
}
```

No omitir validación:

```go
if err := uc.CrearPersona(ctx, payload); err != nil {
	return answer.Err(c, err)
}
```

No responder con `c.JSON`:

```go
return c.JSON(400, err.Error())
```

No validar HTTP dentro del usecase:

```go
func (uc *Usecase) CrearPersona(c echo.Context, payload models.PersonaModel) error
```

---

## Checklist

- [ ] El payload tiene tags `chk`.
- [ ] El handler llama `kcheck.Valid`.
- [ ] La validación ocurre después de `binds`.
- [ ] El error se responde con `answer.Err`.
- [ ] El error se convierte con `errs.BadRequestDirect`.
- [ ] No hay validación manual duplicada.
- [ ] Las reglas de negocio siguen en usecase.