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

## Archivos

### Usecases simples

```text
personas.go
productos.go
```

### Usecases complejos

```text
crear_persona.go
listar_personas.go
actualizar_persona.go
eliminar_personas.go
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

Métodos deben ser específicos y descriptivos:

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

---

## Checklist

- [ ] Se usa español como idioma principal.
- [ ] Los nombres son descriptivos.
- [ ] No hay abreviaciones innecesarias.
- [ ] No hay mezcla de idiomas.
- [ ] Los métodos son específicos.
- [ ] Las interfaces tienen nombres claros.
- [ ] Los archivos siguen la convención.
- [ ] Las variables son claras y consistentes.