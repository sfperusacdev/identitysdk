# Makefile

## Objetivo

Estandarizar los comandos operativos del proyecto.

El proyecto debe tener un `Makefile` en la raíz para ejecutar tareas comunes:

- correr servicio
- generar configuración
- ejecutar migraciones
- crear migraciones
- formatear SQL
- construir imagen Docker
- publicar imagen

---

## Variables

```makefile
version := $(shell cat version)
tag_aws_image := $(shell cat scr_aws_tag)
```

### `version`

Lee la versión del archivo:

```text
version
```

### `tag_aws_image`

Lee el tag base de AWS desde:

```text
scr_aws_tag
```

---

## Ejecutar servicio

```makefile
run:
	@TZ=UTC go run main.go -c config.yml
```

Reglas:

- usar `TZ=UTC`
- usar `main.go`
- usar `-c config.yml`

---

## Generar system props

```makefile
system-props:
	@go run main.go system-props -p systemprops > system_props/constants.go
```

---

## Generar ejemplo de configuración

```makefile
config-example:
	@TZ=UTC go run main.go config-example
```

---

## Migraciones

### Upgrade

```makefile
upgrade:
	@TZ=UTC go run main.go -c config.yml upgrade
```

### Downgrade

```makefile
downgrade:
	@TZ=UTC go run main.go -c config.yml downgrade
```

### Status

```makefile
status:
	@TZ=UTC go run main.go -c config.yml status
```

---

## Crear migración

```makefile
new-migrate:
	@goose -dir=./migrations create $(filter-out $@,$(MAKECMDGOALS)) sql
```

Uso:

```bash
make new-migrate persona
```

Genera una migración en:

```text
migrations/
```

---

## Formatear SQL

```makefile
fmt-sql:
	@sqruff --config .sqruff fix migrations/
```

Reglas:

- usar `sqruff`
- formatear migraciones antes de commit
- usar configuración `.sqruff`

---

## Docker image

```makefile
image:
	docker build --network host -t boletas_api:$(version) .
```

---

## Push AWS

```makefile
push-aws:
	docker tag boletas_api:$(version) $(tag_aws_image):$(version) &&\
	docker push $(tag_aws_image):$(version)
```

---

## Template estándar

```makefile
version := $(shell cat version)
tag_aws_image := $(shell cat scr_aws_tag)

run:
	@TZ=UTC go run main.go -c config.yml

system-props:
	@go run main.go system-props -p systemprops > system_props/constants.go

config-example:
	@TZ=UTC go run main.go config-example

upgrade:
	@TZ=UTC go run main.go -c config.yml upgrade

downgrade:
	@TZ=UTC go run main.go -c config.yml downgrade

status:
	@TZ=UTC go run main.go -c config.yml status

new-migrate:
	@goose -dir=./migrations create $(filter-out $@,$(MAKECMDGOALS)) sql

fmt-sql:
	@sqruff --config .sqruff fix migrations/

image:
	docker build --network host -t <service_name>:$(version) .

push-aws:
	docker tag <service_name>:$(version) $(tag_aws_image):$(version) &&\
	docker push $(tag_aws_image):$(version)
```

---

## Reglas obligatorias

- Todo proyecto debe tener `Makefile`.
- Usar `TZ=UTC` para comandos de ejecución y migraciones.
- Usar `config.yml` en comandos operativos.
- Usar `goose` para crear migraciones.
- Usar `sqruff` para formatear SQL.
- La imagen Docker debe taggearse con `version`.
- No hardcodear versión en el Makefile.

---

## Anti-patterns

No crear migraciones manualmente con nombres inválidos:

```bash
touch migrations/persona.sql
```

No correr migraciones sin config:

```bash
go run main.go upgrade
```

No omitir `TZ=UTC`:

```bash
go run main.go -c config.yml upgrade
```

No hardcodear versión:

```makefile
version := 0.0.1
```

---

## Checklist

- [ ] Existe `Makefile`.
- [ ] Lee `version` desde archivo.
- [ ] Tiene `run`.
- [ ] Tiene `upgrade`.
- [ ] Tiene `downgrade`.
- [ ] Tiene `status`.
- [ ] Tiene `new-migrate`.
- [ ] Tiene `fmt-sql`.
- [ ] Tiene `image`.
- [ ] Usa `TZ=UTC`.
- [ ] Usa `goose`.
- [ ] Usa `sqruff`.