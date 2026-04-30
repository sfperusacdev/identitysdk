# Validación con kcheck

## Objetivo

Estandarizar la validación estructural de structs usando tags `chk`.

La librería estándar para validación es:

```go
"github.com/user0608/goones/kcheck"
```

---

## Uso de kcheck

El uso de `kcheck` no es obligatorio en todos los endpoints.

Debe utilizarse cuando sea necesario validar:

- campos requeridos
- formato
- longitud
- rangos
- reglas simples de estructura

---

## Regla de decisión

```text
validación estructural → handler
validación de negocio  → usecase
```

---

## Handler

El handler es responsable de la validación estructural del request.

Ejemplo:

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

## Usecase

El usecase NO debe validar campos básicos del payload.

No validar en usecase:

- campos requeridos
- strings vacíos
- formatos
- tipos
- longitudes
- rangos simples

Incorrecto:

```go
if producto.Codigo == "" {
	return errs.BadRequestDirect("codigo requerido")
}
```

```go
if producto.Nombre == "" {
	return errs.BadRequestDirect("nombre requerido")
}
```

El usecase solo debe validar reglas de negocio.

Ejemplos:

- existencia en base de datos
- permisos
- estados permitidos
- reglas cruzadas entre entidades
- validaciones contra servicios externos

---

## Cuándo usar kcheck

Usar `kcheck` cuando:

- el payload tiene validaciones declarativas claras
- se requiere consistencia en validación
- se desea evitar validación manual repetitiva

---

## Cuándo NO usar kcheck

No es necesario usar `kcheck` cuando:

- el payload no requiere validación
- el endpoint no recibe body/query estructurado
- la validación depende de lógica de negocio
- la validación requiere DB o servicios externos

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

- Usar `kcheck` solo cuando aporte valor.
- Validar payloads estructurales en el handler.
- Validar después de `binds.JSON` o `binds.Query`.
- Responder errores de validación con:

```go
return answer.Err(c, errs.BadRequestDirect(err.Error()))
```

- Usar tags `chk` en `models` o `dtos` cuando aplique.
- No duplicar validaciones manuales si existe un tag `chk`.
- No usar validadores externos si `kcheck` cubre el caso.
- No validar campos básicos dentro del usecase.
- Usar `dtos` solo cuando el modelo no represente bien la entrada o salida.

---

## Anti-patterns

No validar campos básicos en usecase:

```go
if producto.Nombre == "" {
	return errs.BadRequestDirect("nombre requerido")
}
```

No validar manualmente si existe tag:

```go
if payload.Email == "" {
	return answer.Err(c, errs.BadRequestDirect("email requerido"))
}
```

No omitir validación cuando el payload la requiere:

```go
if err := uc.CrearPersona(ctx, payload); err != nil {
	return answer.Err(c, err)
}
```

No responder con `c.JSON`:

```go
return c.JSON(400, err.Error())
```

No pasar HTTP al usecase:

```go
func (uc *Usecase) CrearPersona(c echo.Context, payload models.PersonaModel) error
```

---

## Checklist

- [ ] Se usa `kcheck` solo cuando aplica.
- [ ] La validación estructural ocurre en handler.
- [ ] El usecase no valida campos básicos.
- [ ] El usecase solo valida reglas de negocio.
- [ ] La validación ocurre después de `binds`.
- [ ] El error se responde con `answer.Err`.
- [ ] El error se convierte con `errs.BadRequestDirect`.
- [ ] No hay validación manual duplicada.