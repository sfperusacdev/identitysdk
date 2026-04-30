# Recipe: Create Migration

## Objetivo

Crear una migración SQL usando `goose` siguiendo el estándar del proyecto.

---

## Ubicación

Las migraciones deben crearse en:

```text
migrations/
```

---

## Naming del archivo

Formato obligatorio:

```text
YYYYMMDDHHMMSS_nombre_descriptivo.sql
```

Ejemplo:

```text
20250210151612_persona.sql
```

Reglas:

- usar timestamp `YYYYMMDDHHMMSS`
- usar `snake_case`
- usar minúsculas
- usar nombre descriptivo
- no usar espacios
- no modificar migraciones existentes

---

## Estructura base

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

## Reglas para tablas principales

Toda tabla principal debe tener:

```sql
codigo varchar(100) not null primary key
```

Y campos base:

```sql
created_by varchar(100) not null,
created_at timestamp without time zone not null,
updated_by varchar(100) not null,
updated_at timestamp without time zone not null
```

---

## Reglas de códigos

Los códigos:

- vienen del frontend
- no son autogenerados
- no son autoincrementales
- se guardan con prefijo multitenant

No usar:

```sql
id serial primary key
```

```sql
id bigserial primary key
```

---

## Tablas intermedias

Para relaciones muchos-a-muchos, se permite omitir auditoría si no aporta valor.

```sql
create table persona_roles
(
    persona_codigo varchar(100) not null,
    rol_codigo varchar(100) not null,
    primary key (persona_codigo, rol_codigo)
);
```

---

## Prohibido

No insertar datos de dominio:

```sql
insert into persona (...) values (...);
```

No crear seeds funcionales.

No incluir lógica de negocio.

---

## Checklist antes de finalizar

- [ ] Archivo está en `migrations/`
- [ ] Nombre usa `YYYYMMDDHHMMSS_nombre.sql`
- [ ] Usa formato goose
- [ ] Tiene `Up`
- [ ] Tiene `Down`
- [ ] `Down` revierte `Up`
- [ ] Tabla principal usa `codigo`
- [ ] Tabla principal tiene auditoría
- [ ] No usa autoincrementales
- [ ] No contiene inserts de dominio
- [ ] No modifica migraciones existentes

---

## Prompt recomendado para IA

```text
Crea una migración goose siguiendo `.ai/recipes/create-migration.md` y `.ai/standards/migrations.md`.

El archivo debe estar en `migrations/`.
Usa nombre `YYYYMMDDHHMMSS_nombre_descriptivo.sql`.
Solo crea estructura de base de datos.
No agregues inserts.
No uses autoincrementales.
Las tablas principales usan `codigo varchar(100) not null primary key`.
Las tablas principales incluyen created_by, created_at, updated_by y updated_at.
```