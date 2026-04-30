# Project Structure

## Objetivo

Definir la estructura estándar de los proyectos para garantizar:

- consistencia entre repositorios
- claridad en responsabilidades
- compatibilidad con generación de código por IA
- facilidad de mantenimiento

---

## Estructura base

```text
.
├── main.go
├── go.mod
├── go.sum
├── version
├── config.yaml
├── migrations/
├── internal/
│   └── <context-name>/
│       ├── domain/
│       │   ├── <usecase>.go
│       │   ├── models/
│       │   ├── dtos/
│       │   ├── repositories/
│       │   └── services/
│       ├── handlers/
│       ├── repos/
│       ├── servs/
│       └── module.go
└── .ai/
```

---

## Raíz del proyecto

### `main.go`

Punto de entrada del servicio.

Debe usar `identitysdk/setup`.

No debe contener lógica de negocio.

---

### `version`

Archivo embebido con la versión del servicio.

Ejemplo:

```text
0.0.1
```

---

### `config.yaml`

Configuración estándar del servicio.

Debe contener:

```yaml
address: "0.0.0.0:7722"
identity: "http://..."
identity_access_token: "..."
database:
  host: "..."
  port: 5432
  db_name: "..."
  username: "..."
  password: "..."
```

---

### `migrations/`

Carpeta para migraciones de base de datos.

Debe existir aunque esté vacía:

```text
migrations/.keep
```

---

## Carpeta `internal/`

Contiene toda la lógica del sistema.

Se organiza por contextos (bounded contexts).

---

## Contextos

Cada contexto representa un dominio funcional.

Ejemplo:

```text
internal/almacen/
internal/ventas/
internal/usuarios/
```

---

## Estructura de un contexto

```text
internal/<context-name>/

├── domain/
├── handlers/
├── repos/
├── servs/
└── module.go
```

---

## `domain/`

Contiene la lógica de negocio.

```text
domain/
├── <usecase>.go
├── models/
├── dtos/
├── repositories/
└── services/
```

---

### `domain/models/`

Modelos principales del dominio.

Se usan en:

- usecases
- repositories

---

### `domain/dtos/`

Estructuras específicas de entrada/salida.

Usar solo cuando:

- el modelo no representa bien el caso
- reportes
- agregaciones

---

### `domain/repositories/`

Interfaces para acceso a base de datos.

No contiene implementación.

---

### `domain/services/`

Interfaces para servicios externos.

No contiene implementación.

---

### `domain/<usecase>.go`

Contiene los usecases.

Puede ser:

#### Simple

```text
personas.go
```

#### Complejo

```text
crear_persona.go
listar_personas.go
actualizar_persona.go
```

---

## `handlers/`

Contiene handlers HTTP.

Responsabilidades:

- parsear request (`binds`)
- validar (`kcheck`)
- aplicar multitenancy (`identitysdk`)
- llamar usecase
- responder (`answer`)

No contiene lógica de negocio.

---

## `repos/`

Implementación de interfaces de:

```text
domain/repositories
```

Responsable de:

- acceso a base de datos
- queries
- transacciones

---

## `servs/`

Implementación de interfaces de:

```text
domain/services
```

Responsable de:

- integraciones externas
- llamadas a APIs
- lógica fuera del sistema

---

## `module.go`

Archivo de integración con Fx.

Ejemplo:

```go
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

---

## `.ai/`

Contiene los estándares y reglas del proyecto.

```text
.ai/
├── recipes/
├── patterns/
└── standards/
```

---

## Reglas obligatorias

- Toda lógica debe vivir dentro de `internal/`.
- No crear paquetes fuera de `internal/` para lógica de negocio.
- Separar por contexto.
- No mezclar responsabilidades entre carpetas.
- Interfaces en `domain`, implementaciones fuera.
- Handlers no contienen lógica de negocio.
- Usecases no contienen lógica HTTP.
- Repositories no contienen lógica de negocio.
- Services no acceden a base de datos.

---

## Anti-patterns

No mezclar capas:

```text
internal/handlers/db.go
```

No poner lógica en `main.go`:

```go
func main() {
	// lógica de negocio ❌
}
```

No implementar interfaces dentro de `domain/`:

```text
domain/repositories/impl.go ❌
```

No usar una sola carpeta para todo:

```text
internal/
  handlers/
  repos/
  models/
```

No ignorar contextos:

```text
internal/
  personas/
  productos/
```

sin estructura interna.

---

## Checklist

- [ ] Existe `internal/`.
- [ ] Existe al menos un contexto.
- [ ] Cada contexto tiene `domain`, `handlers`, `repos`, `servs`, `module.go`.
- [ ] `domain` contiene modelos, dtos e interfaces.
- [ ] `repos` implementa repositories.
- [ ] `servs` implementa services.
- [ ] `handlers` define rutas HTTP.
- [ ] `module.go` registra todo con Fx.
- [ ] No hay lógica fuera de `internal/`.
- [ ] `.ai/` existe con estándares.