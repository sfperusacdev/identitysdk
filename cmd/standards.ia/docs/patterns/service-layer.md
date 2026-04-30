# Service Layer / Usecases

## Objetivo

Definir cómo se implementa la lógica de negocio dentro de un contexto del sistema.

El usecase es responsable de:

- aplicar reglas de negocio
- orquestar repositorios
- orquestar servicios externos
- propagar `context.Context`
- retornar errores estándar

---

## Estructura

```text
internal/<context-name>/

  domain/
    <usecase>.go
    models/
    dtos/
    repositories/
    services/

  handlers/
  repos/
  servs/
  module.go
```

---

## Responsabilidad de cada carpeta

### `domain/`

Contiene la lógica de negocio del contexto.

Aquí viven los usecases.

---

### `domain/models/`

Contiene modelos principales del dominio.

En CRUDs simples, usar directamente `models`.

```go
models.PersonaModel
```

---

### `domain/dtos/`

Contiene estructuras especiales de entrada o salida.

Usar `dtos` solo cuando el modelo no cubre bien el caso, por ejemplo:

- reportes
- queries agregadas
- respuestas calculadas
- estructuras que no representan una entidad principal

No crear DTOs innecesarios si `models` es suficiente.

---

### `domain/repositories/`

Contiene interfaces para acceder a base de datos.

```go
type PersonasRepository interface {
	CrearPersona(ctx context.Context, persona models.PersonaModel) error
	ListarPersonas(ctx context.Context, prefix string) ([]models.PersonaModel, error)
}
```

---

### `domain/services/`

Contiene interfaces para servicios externos o acciones fuera de la base de datos.

```go
type ValidadorSunatService interface {
	ValidarDoc(ctx context.Context, documento string) error
}
```

No usar `domain/services` para acceso a base de datos.

---

## Patrón de usecase

```go
package domain

import (
	"context"

	"app/internal/<context-name>/domain/models"
	"app/internal/<context-name>/domain/repositories"
	"app/internal/<context-name>/domain/services"

	"github.com/user0608/goones/errs"
)

type PersonasUsecase struct {
	repository   repositories.PersonasRepository
	sunatService services.ValidadorSunatService
}

func NewPersonasUsecase(
	repository repositories.PersonasRepository,
	sunatService services.ValidadorSunatService,
) *PersonasUsecase {
	return &PersonasUsecase{
		repository:   repository,
		sunatService: sunatService,
	}
}

func (uc *PersonasUsecase) CrearPersona(ctx context.Context, persona models.PersonaModel) error {
	if persona.Nombre == "" {
		return errs.BadRequestDirect("no se encontro el nombre")
	}

	if err := uc.sunatService.ValidarDoc(ctx, "valida el documento persona"); err != nil {
		return err
	}

	return uc.repository.CrearPersona(ctx, persona)
}

func (uc *PersonasUsecase) ListarPersonas(ctx context.Context, prefix string) ([]models.PersonaModel, error) {
	personas, err := uc.repository.ListarPersonas(ctx, prefix)
	if err != nil {
		return nil, err
	}

	return personas, nil
}
```

---

## Constructor

Todo usecase debe tener constructor explícito.

```go
func NewPersonasUsecase(
	repository repositories.PersonasRepository,
	sunatService services.ValidadorSunatService,
) *PersonasUsecase
```

El constructor debe recibir interfaces, no implementaciones concretas.

Correcto:

```go
func NewPersonasUsecase(repository repositories.PersonasRepository) *PersonasUsecase
```

Incorrecto:

```go
func NewPersonasUsecase(repository *repos.PersonaRepository) *PersonasUsecase
```

---

## Dependencias

El usecase solo puede depender de:

```text
domain/models
domain/dtos
domain/repositories
domain/services
```

También puede usar librerías transversales permitidas como:

```go
"github.com/user0608/goones/errs"
```

---

## Reglas para interfaces

Las interfaces deben tener nombres descriptivos.

Correcto:

```go
type PersonasRepository interface {}
type ValidadorSunatService interface {}
```

Incorrecto:

```go
type Repository interface {}
type Service interface {}
type Interface interface {}
```

---

## Reglas para métodos

Los métodos deben tener nombres descriptivos y específicos.

Correcto:

```go
CrearPersona
ListarPersonas
ActualizarPersona
EliminarPersonas
ValidarDoc
```

Incorrecto:

```go
Crear
Listar
Actualizar
Eliminar
Do
Run
Handle
```

---

## Uso de `context.Context`

Todos los métodos públicos del usecase deben recibir `context.Context` como primer parámetro.

Correcto:

```go
func (uc *PersonasUsecase) CrearPersona(ctx context.Context, persona models.PersonaModel) error
```

Incorrecto:

```go
func (uc *PersonasUsecase) CrearPersona(persona models.PersonaModel) error
```

El mismo `ctx` recibido desde el handler debe propagarse hacia repositories y services.

---

## Multitenancy

El handler es responsable de preparar los datos multitenant.

El usecase debe asumir que:

- los códigos/IDs recibidos desde el cliente ya pasaron por `identitysdk.Empresa`
- los listados reciben `prefix` calculado con `identitysdk.EmpresaPrefix`

No agregar `%` manualmente al `prefix`.

Correcto:

```go
prefix := identitysdk.EmpresaPrefix(ctx)
```

Incorrecto:

```go
prefix := identitysdk.EmpresaPrefix(ctx) + "%"
```

Ejemplo para listado:

```go
func (uc *PersonasUsecase) ListarPersonas(ctx context.Context, prefix string) ([]models.PersonaModel, error) {
	return uc.repository.ListarPersonas(ctx, prefix)
}
```

El usecase no debe calcular el prefijo por su cuenta.

---

## Servicios externos

Si una operación requiere comunicarse con otro sistema, usar interfaces en:

```text
domain/services/
```

Ejemplo:

```go
type ValidadorSunatService interface {
	ValidarDoc(ctx context.Context, documento string) error
}
```

El usecase llama la interfaz:

```go
if err := uc.sunatService.ValidarDoc(ctx, documento); err != nil {
	return err
}
```

La implementación vive en:

```text
servs/
```

---

## Repositorios

Si una operación requiere base de datos, usar interfaces en:

```text
domain/repositories/
```

Ejemplo:

```go
type PersonasRepository interface {
	CrearPersona(ctx context.Context, persona models.PersonaModel) error
	ListarPersonas(ctx context.Context, prefix string) ([]models.PersonaModel, error)
}
```

La implementación vive en:

```text
repos/
```

---

## Separación por complejidad

Para CRUDs simples se puede usar un solo archivo:

```text
domain/personas.go
```

Para lógica más avanzada, separar por operación:

```text
domain/
  crear_persona.go
  listar_personas.go
  actualizar_persona.go
  eliminar_personas.go
```

---

## Errores

Para errores de negocio usar `errs`.

Ejemplo:

```go
return errs.BadRequestDirect("no se encontro el nombre")
```

Los errores recibidos desde repositories o services deben propagarse.

```go
if err := uc.repository.CrearPersona(ctx, persona); err != nil {
	return err
}
```

---

## Responsabilidades permitidas

El usecase puede:

- validar reglas de negocio
- orquestar repositories
- orquestar services
- transformar modelos o DTOs
- decidir el flujo de negocio
- retornar errores de negocio

---

## Responsabilidades prohibidas

El usecase no debe:

- usar `echo.Context`
- usar `answer`
- usar `binds`
- parsear JSON
- leer query params
- responder HTTP
- acceder directamente a DB
- abrir conexiones DB
- crear dependencias manualmente
- usar implementaciones concretas de `repos`
- usar implementaciones concretas de `servs`
- usar `context.Background()` dentro de un flujo HTTP

---

## Anti-patterns

No usar HTTP en usecase:

```go
func (uc *PersonasUsecase) CrearPersona(c echo.Context) error
```

No responder desde usecase:

```go
return answer.Ok(c, persona)
```

No acceder a DB directamente:

```go
db.Query(...)
```

No depender de implementación concreta:

```go
type PersonasUsecase struct {
	repository *repos.PersonaRepository
}
```

No crear dependencias manualmente:

```go
repo := repos.NewPersonaRepository(manager)
```

No usar nombres genéricos:

```go
type Interface interface{}
func (uc *Usecase) Listar(ctx context.Context) error
```

---

## Checklist

- [ ] El usecase está en `internal/<context-name>/domain`.
- [ ] Usa `domain/models` para entidades principales.
- [ ] Usa `domain/dtos` solo cuando el modelo no cubre el caso.
- [ ] Depende de interfaces en `domain/repositories`.
- [ ] Depende de interfaces en `domain/services` cuando aplica.
- [ ] No depende de `repos`.
- [ ] No depende de `servs`.
- [ ] Tiene constructor explícito.
- [ ] El constructor recibe interfaces.
- [ ] Todos los métodos reciben `context.Context`.
- [ ] Los métodos tienen nombres descriptivos.
- [ ] No usa HTTP.
- [ ] No usa `answer`.
- [ ] No usa `binds`.
- [ ] No accede directamente a DB.
- [ ] No calcula tenant por su cuenta.
- [ ] Propaga errores correctamente.