# Naming

## Objetivo

Estandarizar los nombres usados en el proyecto para garantizar:

- consistencia
- claridad
- facilidad de mantenimiento
- correcta generación de código por IA

---

## Idioma

El idioma oficial del código es **español**.

---

## Reglas generales

- Usar nombres descriptivos.
- Evitar abreviaciones innecesarias.
- Mantener consistencia en todo el proyecto.
- Usar PascalCase para tipos y funciones públicas.
- Usar camelCase para variables.

---

## Usecases

### Nombre del struct

```go
PersonasUsecase
ProductosUsecase
UsuariosUsecase
```

### Métodos

```go
CrearPersona
ListarPersonas
ActualizarPersona
EliminarPersonas
```

---

## Repositories

### Nombre del struct

```go
PersonaRepository
ProductoRepository
UsuarioRepository
```

### Interfaces

```go
type PersonasRepository interface {}
```

### Métodos

```go
CrearPersona
ListarPersonas
ActualizarPersona
EliminarPersonas
```

---

## Services

### Nombre del struct

```go
ValidadorSunatService
NotificadorEmailService
IntegracionPagoService
```

### Interfaces

```go
type ValidadorSunatService interface {}
```

### Métodos

```go
ValidarDoc
EnviarCorreo
ProcesarPago
```

---

## Handlers

### Nombre de funciones

```go
CrearPersonaHandler
ListarPersonaHandler
ActualizarPersonaHandler
EliminarPersonaHandler
```

---

## Models

```go
PersonaModel
ProductoModel
UsuarioModel
```

---

## DTOs

Usar solo cuando sea necesario.

```go
CrearPersonaRequest
ActualizarPersonaRequest
PersonaResponse
ReporteVentasDTO
```

---

## Naming de archivos

Los archivos NO deben usar sufijos redundantes como:

- _model
- _repository
- _service
- _handler

El nombre del archivo debe ser simple y representar el dominio o acción.

---

### Ejemplos correctos

```text
personas.go
persona.go
productos.go
almacenes.go
```

Casos específicos:

```text
crear_persona.go
listar_personas.go
actualizar_persona.go
eliminar_personas.go
```

---

### Ejemplos incorrectos

```text
persona_model.go
persona_repository.go
persona_service.go
persona_handler.go
```

---

## Relación archivo → contenido

```text
archivo                      contenido
---------------------------------------------
personas.go                 PersonasUsecase
persona.go                  PersonaModel
repos/persona.go            PersonaRepository
servs/sunat.go              ValidadorSunatService
handlers/persona.go         handlers HTTP
```

---

## Variables

```go
persona
personas
repo
service
ctx
err
```

---

## Parámetros comunes

```go
ctx context.Context
prefix string
codigo string
codigos []string
```

---

## Naming de contextos

```text
almacen
ventas
usuarios
facturacion
```

No usar:

```text
modulo1
test
temp
misc
```

---

## Naming en interfaces

Interfaces deben ser específicas:

```go
type PersonasRepository interface {}
type ValidadorSunatService interface {}
```

No usar nombres genéricos:

```go
type Repository interface {}
type Service interface {}
type Interface interface {}
```

---

## Naming en métodos

Métodos deben ser específicos y descriptivos.

Correcto:

```go
CrearPersona
ListarPersonas
ActualizarPersona
EliminarPersonas
```

Incorrecto:

```go
Create
List
Do
Run
Handle
Process
Execute
```

---

## Anti-patterns

No usar nombres genéricos:

```go
func Do() {}
func Run() {}
func Handle() {}
```

No usar abreviaciones innecesarias:

```go
usr
prd
svc
```

No mezclar idiomas:

```go
CrearUser
DeletePersona
ListUsuarios
```

No usar nombres ambiguos:

```go
Manager
Helper
Util
Common
```

No duplicar contexto en nombres de archivos:

```text
persona_repository.go
persona_model.go
```

---

## Checklist

- [ ] Se usa español como idioma principal
- [ ] Los nombres son descriptivos
- [ ] No hay abreviaciones innecesarias
- [ ] No hay mezcla de idiomas
- [ ] Los métodos son específicos
- [ ] Las interfaces tienen nombres claros
- [ ] Los archivos NO usan sufijos redundantes
- [ ] Los nombres de archivos son simples y claros
- [ ] Las variables son consistentes