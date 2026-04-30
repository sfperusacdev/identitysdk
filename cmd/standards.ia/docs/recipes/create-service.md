# Crear servicio

## Objetivo

Crear un servicio Go usando el bootstrap estándar de la empresa basado en:

```go
github.com/sfperusacdev/identitysdk/setup
```

Este bootstrap ya incluye las librerías internas base de la empresa. Al ejecutar `service.Run(...)`, el servicio levanta automáticamente el servidor HTTP usando la configuración estándar.

---

## Estructura obligatoria

```text
main.go
migrations/.keep
version
config.yaml
```

---

## Archivo `version`

Para la primera versión del servicio, crear el archivo:

```text
version
```

Con el contenido:

```text
0.0.1
```

---

## Directorio `migrations`

Crear siempre el directorio:

```text
migrations/
```

Si el servicio aún no tiene migraciones, agregar:

```text
migrations/.keep
```

Esto permite versionar el directorio vacío en Git.

---

## Archivo de configuración

Crear el archivo:

```text
config.yaml
```

Debe seguir la estructura definida en:

```text
.ai/standards/configuration.md
```

El puerto HTTP del servicio NO se define en código. Se define en el campo:

```yaml
address: "0.0.0.0:<port>"
```

---

## `service_id`

`service_id` es el identificador único del servicio y tiene la siguiente estructura com.sfperusac.<service_name>_api

Se usa para identificar o registrar el servicio contra el servicio de identidad interno de la empresa.

Debe ser estable, único y descriptivo.

Ejemplo:

```go
setup.WithDetails("com.sfperusac.tareo_api", "Servicio de gestión de cursos")
```

---

## `service_description`

`service_description` es una descripción corta del propósito del servicio.

Debe explicar qué responsabilidad tiene el servicio dentro del sistema.

---

## `main.go` obligatorio

Ruta:

```text
main.go
```

Contenido base:

```go
package main

import (
	"embed"
	"log/slog"
	"os"

	"github.com/sfperusacdev/identitysdk/setup"
)

//go:embed version
var version string

//go:embed all:migrations
var migrationsDir embed.FS

func main() {
	service := setup.NewService(version,
		setup.WithDetails("<service_id>", "<service_description>"),
		setup.WithMigrationSource(migrationsDir),
	)

	err := service.Run(
		// fx.Option
	)
	if err != nil {
		slog.Error("Failed to run the service", "error", err)
		os.Exit(1)
	}
}
```

---

## Reglas obligatorias

- Siempre usar `setup.NewService`.
- Siempre usar `setup.WithDetails`.
- Siempre usar `setup.WithMigrationSource`.
- Siempre embeder el archivo `version`.
- Siempre embeder el directorio `migrations`.
- Siempre crear `version` con `0.0.1` en la primera versión.
- Siempre crear `migrations/.keep` si no existen migraciones iniciales.
- Siempre leer el puerto HTTP desde `config.yaml`.
- No crear un servidor HTTP manualmente.
- No usar `http.ListenAndServe` directamente.
- No hardcodear puertos en código.
- No inicializar manualmente librerías internas que ya son provistas por `identitysdk/setup`.
- No omitir `service_id`.
- No omitir `service_description`.

---

## Anti-patterns

No levantar HTTP manualmente:

```go
http.ListenAndServe(":8080", handler)
```

No hardcodear el puerto:

```go
port := "8080"
```

No crear un bootstrap alternativo:

```go
service := MyCustomService{}
```

No omitir migraciones:

```text
# incorrecto
# no existe migrations/
```

No omitir version:

```text
# incorrecto
# no existe version
```

---

## Checklist

Antes de finalizar un nuevo servicio, verificar:

- [ ] Existe `main.go`.
- [ ] Existe `version`.
- [ ] `version` contiene `0.0.1` en la primera versión.
- [ ] Existe `migrations/`.
- [ ] Existe `migrations/.keep` si no hay migraciones.
- [ ] Existe `config.yaml`.
- [ ] `config.yaml` sigue `.ai/standards/configuration.md`.
- [ ] `main.go` usa `setup.NewService`.
- [ ] `main.go` usa `setup.WithDetails`.
- [ ] `main.go` usa `setup.WithMigrationSource`.
- [ ] El puerto HTTP viene desde `config.yaml`.
- [ ] No se crea servidor HTTP manualmente.
- [ ] No se hardcodean puertos en código.
- [ ] No se inicializan manualmente librerías internas ya incluidas por `identitysdk/setup`.

---

## Prompt recomendado para IA

Cuando se use una IA para crear un nuevo servicio, indicar:

```text
Crea un nuevo servicio Go siguiendo estrictamente `.ai/recipes/create-service.md`.
Usa el bootstrap estándar con `identitysdk/setup`.
No crees un servidor HTTP manual.
El puerto debe venir desde `config.yaml`.
Crea `version` con `0.0.1`.
Crea `migrations/.keep` si no hay migraciones iniciales.
```