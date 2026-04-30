# Repository Layer

## Objetivo

Definir cómo se implementa el acceso a datos en la aplicación.

El repository es responsable de:

- interactuar con la base de datos
- ejecutar queries
- mapear resultados a modelos
- manejar errores de base de datos
- manejar transacciones

---

## Ubicación

```text
internal/<context-name>/repos/
```

Ejemplo:

```text
internal/almacen/repos/
```

---

## Relación con domain

```text
domain/repositories → interfaces
repos/              → implementación
```

El repository implementa interfaces definidas en:

```text
internal/<context-name>/domain/repositories/
```

---

## Estructura básica

```go
package repos

import (
	"context"

	"app/internal/<context-name>/domain/models"
	"app/internal/<context-name>/domain/repositories"

	connection "github.com/sfperusacdev/identitysdk/pg-connection"
	"github.com/user0608/goones/errs"
)

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

## Reglas obligatorias

- El repository debe implementar una interface de `domain/repositories`.
- Validar implementación con:

```go
var _ repositories.PersonasRepository = (*PersonaRepository)(nil)
```

- El repository debe recibir `connection.StorageManager`.
- Nunca abrir conexiones manualmente.
- Siempre usar `context.Context`.
- Nunca usar `context.Background()`.

---

## Uso de conexión

Siempre obtener conexión así:

```go
tx := r.manager.Conn(ctx)
```

Nunca:

```go
db.Open(...)
```

---

## Ejemplo: Crear

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

## Ejemplo: Listar con multitenancy

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

## Multitenancy

Para consultas por empresa, usar siempre:

```go
prefix string
```

Este valor proviene del handler:

```go
identitysdk.EmpresaPrefix(ctx)
```

Ejemplo:

```sql
WHERE codigo LIKE ?
```

---

## Regla crítica

Nunca construir el prefix manualmente.

Correcto:

```go
prefix := identitysdk.EmpresaPrefix(ctx)
```

Incorrecto:

```go
prefix := identitysdk.EmpresaPrefix(ctx) + "%"
```

```go
prefix := empresa + "%"
```

---

## Manejo de errores

Todos los errores de base de datos deben envolverse con:

```go
errs.Pgf(err)
```

Ejemplo:

```go
if rs.Error != nil {
	return errs.Pgf(rs.Error)
}
```

---

## Uso de `rs` (resultado de query)

Siempre conservar la variable `rs`:

```go
rs := tx.Raw(qry, prefix).Scan(&personas)
if rs.Error != nil {
	return nil, errs.Pgf(rs.Error)
}
```

---

## Regla importante

No separar `err` directamente desde la query:

Incorrecto:

```go
err := tx.Raw(qry, prefix).Scan(&personas).Error
if err != nil {
	return nil, errs.Pgf(err)
}
```

Motivo:

- se pierde consistencia
- dificulta debugging
- rompe el patrón estándar del proyecto

---

## Transacciones

Para operaciones que requieren transacción, usar:

```go
err := r.manager.WithTx(ctx, func(ctx context.Context) error {
	tx := r.manager.Conn(ctx)

	// operaciones dentro de la transacción

	return nil
})
if err != nil {
	return errs.Pgf(err)
}
```

---

## Comportamiento de `Conn(ctx)`

Dentro de una transacción:

```go
tx := r.manager.Conn(ctx)
```

retorna la transacción activa, no una conexión nueva.

---

## Reglas de transacción

- Usar `r.manager.WithTx(ctx, func(ctx context.Context) error { ... })`.
- No abrir transacciones manualmente.
- No usar `Begin`, `Commit` o `Rollback`.
- Dentro del callback, usar el `ctx` recibido.
- Dentro del callback, obtener conexión con `r.manager.Conn(ctx)`.
- Si ocurre error, retornar el error para hacer rollback.
- Si todo está correcto, retornar `nil`.

---

## Ejemplo con transacción

```go
func (r *PersonaRepository) CrearPersonas(ctx context.Context, personas []models.PersonaModel) error {
	err := r.manager.WithTx(ctx, func(ctx context.Context) error {
		tx := r.manager.Conn(ctx)

		for _, persona := range personas {
			rs := tx.Table("persona").Create(&persona)
			if rs.Error != nil {
				return errs.Pgf(rs.Error)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
```

---

## Naming

### Struct

```go
PersonaRepository
```

### Métodos

```go
CrearPersona
ListarPersonas
EliminarPersonas
```

Evitar:

```go
Repo
Create
List
Do
Run
```

---

## Responsabilidades

El repository puede:

- ejecutar queries SQL
- usar ORM
- mapear resultados
- manejar errores de DB
- ejecutar transacciones

---

## Responsabilidades prohibidas

El repository no debe:

- contener lógica de negocio
- usar `echo.Context`
- usar `answer`
- usar `binds`
- usar `identitysdk`
- crear conexiones DB manualmente
- acceder a HTTP
- llamar a handlers

---

## Anti-patterns

No abrir conexión manual:

```go
db, _ := sql.Open(...)
```

No usar contexto:

```go
r.manager.Conn(context.Background())
```

No construir prefix manual:

```go
prefix := identitysdk.EmpresaPrefix(ctx) + "%"
```

No usar err directo:

```go
err := tx.Raw(qry).Scan(&data).Error
```

No ignorar errores:

```go
tx.Raw(qry).Scan(&data)
```

No acceder a DB desde handler:

```go
db.Query(...)
```

No mezclar lógica de negocio:

```go
if persona.Nombre == "" {
	return error
}
```

No usar transacciones manuales:

```go
tx := db.Begin()
tx.Commit()
tx.Rollback()
```

No usar contexto incorrecto en transacción:

```go
err := r.manager.WithTx(ctx, func(txCtx context.Context) error {
	tx := r.manager.Conn(ctx) // incorrecto
	return nil
})
```

---

## Checklist

- [ ] Ubicado en `internal/<context-name>/repos`
- [ ] Implementa interface de `domain/repositories`
- [ ] Tiene validación `var _ interface = (*struct)(nil)`
- [ ] Usa `StorageManager`
- [ ] Usa `ctx`
- [ ] No usa `context.Background`
- [ ] Usa `errs.Pgf` para errores
- [ ] Usa `rs := ...` para queries
- [ ] No usa `err := query.Error`
- [ ] No construye `%` manualmente
- [ ] Usa `WithTx` para transacciones
- [ ] No usa Begin/Commit/Rollback
- [ ] No contiene lógica de negocio
- [ ] No accede a HTTP
- [ ] Naming descriptivo

---

## Prompt recomendado para IA

```text
Crea un repository siguiendo `.ai/patterns/repository.md`.

Usa StorageManager.
Obtén conexión con manager.Conn(ctx).
Usa rs := ... para queries.
Maneja errores con errs.Pgf.
No uses err := query.Error directamente.
No construyas prefix manualmente.
Usa WithTx para transacciones.
No uses Begin/Commit/Rollback.
Implementa la interface de domain/repositories.
Valida implementación con var _ interface.
No pongas lógica de negocio en el repository.
```