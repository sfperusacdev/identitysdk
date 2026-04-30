# Migrations

## Objetivo

Estandarizar la creación y manejo de migraciones de base de datos.

El proyecto usa `goose` para gestionar migraciones.

---

## Ubicación

```text
migrations/
```

---

## Regla principal

Las migraciones deben contener únicamente estructura de base de datos.

Permitido:

- CREATE TABLE
- ALTER TABLE
- DROP TABLE
- índices
- constraints
- llaves foráneas

No permitido:

- inserts de datos de dominio
- seeds funcionales
- lógica de negocio

---

## Estructura obligatoria

```sql
-- +goose Up
-- +goose StatementBegin
create table persona
(
    codigo varchar(100) not null primary key,

    -- campos del dominio

    created_by varchar(100) not null,
    created_at timestamp without time zone not null,
    updated_by varchar(100) not null,
    updated_at timestamp without time zone not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table persona;
-- +goose StatementEnd
```

---

## Campos base obligatorios

Toda tabla principal debe incluir:

```sql
created_by varchar(100) not null,
created_at timestamp without time zone not null,
updated_by varchar(100) not null,
updated_at timestamp without time zone not null
```

Estos campos son manejados por:

```go
identitysdk.CreateBy(ctx, &entity)
identitysdk.UpdateBy(ctx, &entity)
```

---

## Campo `codigo`

Todas las tablas principales deben usar:

```sql
codigo varchar(100) not null primary key
```

Reglas:

- viene del frontend
- no es autoincremental
- incluye prefijo multitenant
- se transforma con `identitysdk.Empresa`

---

## Tablas intermedias (N:N)

En relaciones muchos-a-muchos:

- se puede omitir auditoría si no aporta valor

Ejemplo:

```sql
create table persona_roles
(
    persona_codigo varchar(100) not null,
    rol_codigo varchar(100) not null,
    primary key (persona_codigo, rol_codigo)
);
```

---

## Relación con modelos Go

Las tablas principales deben tener un modelo que embeba:

```go
identitysdk.Model
```

Ejemplo:

```go
package models

import "github.com/sfperusacdev/identitysdk"

type PersonaModel struct {
	Codigo string `json:"codigo" gorm:"primaryKey" chk:"nonil"`
	Nombre string `json:"nombre"`
	Edad   int    `json:"edad"`

	identitysdk.Model
}
```

Definición:

```go
type Model struct {
	CreatedBy string
	CreatedAt time.Time
	UpdatedBy string
	UpdatedAt time.Time
}
```

---

## Naming de migraciones

Formato:

```text
YYYYMMDDHHMMSS_nombre_descriptivo.sql
```

---

## Ejemplos

```text
20250210151612_persona.sql
20250210153438_lote_etiqueta.sql
20250210154339_preference.sql
20250210155254_documento.sql
```

---

## Reglas de naming

- timestamp obligatorio
- snake_case
- minúsculas
- descriptivo
- único
- orden creciente

---

## Orden de ejecución

`goose` ejecuta por timestamp.

Reglas:

- no modificar migraciones existentes
- no usar timestamps antiguos
- no alterar el orden

---

## Reglas obligatorias

- usar formato goose
- toda migración tiene `Up` y `Down`
- `Down` debe revertir `Up`
- no usar autoincrementales
- no insertar datos de dominio
- no modificar migraciones existentes
- tablas principales deben tener auditoría
- tablas principales deben usar `codigo`

---

## Anti-patterns

No usar autoincrementales:

```sql
id serial primary key
```

No insertar datos:

```sql
insert into persona values (...)
```

No omitir Down:

```sql
-- +goose Down
-- vacío
```

No modificar migraciones:

```text
archivo existente editado ❌
```

No usar nombres genéricos:

```text
test.sql
fix.sql
update.sql
```

No usar mayúsculas o espacios:

```text
Create Persona.sql
```

---

## Checklist

- [ ] Archivo en `migrations/`
- [ ] Usa goose
- [ ] Tiene Up
- [ ] Tiene Down
- [ ] Solo estructura
- [ ] No inserts
- [ ] Tiene codigo
- [ ] Tiene auditoría
- [ ] Naming correcto
- [ ] Orden correcto