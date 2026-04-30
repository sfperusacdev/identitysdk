# Recipe: Create Usecase

## Objetivo

Crear un usecase siguiendo los estándares del proyecto.

---

## Ubicación

```text
internal/<context-name>/domain/
```

---

## Estructura base

```go
package domain

import (
	"context"

	"<module>/internal/<context-name>/domain/models"
	"<module>/internal/<context-name>/domain/repositories"
	"<module>/internal/<context-name>/domain/services"

	"github.com/user0608/goones/errs"
)

type <Nombre>Usecase struct {
	repository   repositories.<NombrePlural>Repository
	serviceExt   services.<NombreServicio>Service
}

func New<Nombre>Usecase(
	repository repositories.<NombrePlural>Repository,
	serviceExt services.<NombreServicio>Service,
) *<Nombre>Usecase {
	return &<Nombre>Usecase{
		repository: repository,
		serviceExt: serviceExt,
	}
}
```

---

## Ejemplo

```go
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
```

---

## Métodos

### Crear

```go
func (uc *PersonasUsecase) CrearPersona(ctx context.Context, persona models.PersonaModel) error {
	if persona.Nombre == "" {
		return errs.BadRequestDirect("nombre requerido")
	}

	if err := uc.sunatService.ValidarDoc(ctx, persona.Codigo); err != nil {
		return err
	}

	return uc.repository.CrearPersona(ctx, persona)
}
```

---

### Listar

```go
func (uc *PersonasUsecase) ListarPersonas(ctx context.Context, prefix string) ([]models.PersonaModel, error) {
	return uc.repository.ListarPersonas(ctx, prefix)
}
```

---

## Reglas obligatorias

- Recibir siempre `context.Context`
- Usar interfaces de `repositories`
- Usar interfaces de `services`
- No usar implementaciones concretas
- No usar `echo.Context`
- No usar `answer`
- No acceder a DB directamente
- No generar códigos
- Propagar errores
- Naming descriptivo

---

## Multitenancy

El usecase:

- recibe IDs ya transformados
- recibe `prefix` en listados
- no usa `identitysdk`

---

## Validación

- Validación estructural → handler
- Validación de negocio → usecase

---

## Anti-patterns

```go
func (uc *Usecase) Crear(c echo.Context) error
```

```go
repo := repos.NewRepo()
```

```go
db.Query(...)
```

```go
context.Background()
```

---

## Checklist

- [ ] Está en `domain/`
- [ ] Usa interfaces
- [ ] Tiene constructor
- [ ] Recibe `ctx`
- [ ] No usa HTTP
- [ ] No usa DB directo
- [ ] Naming correcto