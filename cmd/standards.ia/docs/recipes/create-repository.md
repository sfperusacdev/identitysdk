# Recipe: Create Repository

## Objetivo

Crear un repository siguiendo los estándares del proyecto.

---

## Ubicación

```text
internal/<context-name>/repos/
```

---

## Estructura base

```go
package repos

import (
	"context"

	"<module>/internal/<context-name>/domain/models"
	"<module>/internal/<context-name>/domain/repositories"

	connection "github.com/sfperusacdev/identitysdk/pg-connection"
	"github.com/user0608/goones/errs"
)

type <Nombre>Repository struct {
	manager connection.StorageManager
}

var _ repositories.<NombrePlural>Repository = (*<Nombre>Repository)(nil)

func New<Nombre>Repository(manager connection.StorageManager) *<Nombre>Repository {
	return &<Nombre>Repository{
		manager: manager,
	}
}
```

---

## Ejemplo

```go
type PersonaRepository struct {
	manager connection.StorageManager
}

var _ repositories.PersonasRepository = (*PersonaRepository)(nil)

func NewPersonaRepository(manager connection.StorageManager) *PersonaRepository {
	return &PersonaRepository{
		manager: manager,
	}
}
```

---

## Crear

```go
func (r *PersonaRepository) CrearPersona(ctx context.Context, persona models.PersonaModel) error {
	tx := r.manager.Conn(ctx)

	rs := tx.Table("persona").Create(&persona)
	if rs.Error != nil {
		return errs.Pgf(rs.Error)
	}

	return nil
}
```

---

## Listar

```go
func (r *PersonaRepository) ListarPersonas(ctx context.Context, prefix string) ([]models.PersonaModel, error) {
	tx := r.manager.Conn(ctx)

	var personas []models.PersonaModel
	qry := `select * from persona where codigo like ?`

	rs := tx.Raw(qry, prefix).Scan(&personas)
	if rs.Error != nil {
		return nil, errs.Pgf(rs.Error)
	}

	return personas, nil
}
```

---

## Transacciones

```go
err := r.manager.WithTx(ctx, func(ctx context.Context) error {
	tx := r.manager.Conn(ctx)

	// operaciones

	return nil
})
if err != nil {
	return err
}
```

---

## Reglas obligatorias

- Usar `StorageManager`
- Usar `Conn(ctx)`
- Usar `rs := ...`
- Manejar errores con `errs.Pgf`
- No usar `err := query.Error`
- No abrir DB manualmente
- No usar `context.Background`
- No generar lógica de negocio

---

## Multitenancy

- Usar `prefix` en queries
- No construir `%` manualmente

---

## Anti-patterns

```go
db.Open(...)
```

```go
err := tx.Raw(...).Error
```

```go
prefix + "%"
```

```go
tx := db.Begin()
```

---

## Checklist

- [ ] Está en `repos/`
- [ ] Implementa interface
- [ ] Usa StorageManager
- [ ] Usa ctx
- [ ] Usa errs.Pgf
- [ ] Usa rs :=
- [ ] No lógica de negocio