# Docker

## Objetivo

Estandarizar la construcción de imágenes Docker para los servicios.

---

## Regla base

El nombre del binario y del servicio debe ser dinámico.

No usar nombres hardcodeados como:

```text
tareo
boletas_api
```

Debe usarse el nombre del proyecto.

---

## Estructura estándar

```dockerfile
FROM golang:1.25.5-alpine AS builder

RUN apk add git

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o build/<service_name> *.go

FROM alpine:3.22

RUN apk update && apk upgrade
RUN apk add --update tzdata && rm -rf /var/cache/apk/*

WORKDIR /app

COPY --from=builder /app/build/<service_name> .

CMD ["./<service_name>", "-c", "/etc/<service_name>/config.yml", "--auto"]
```

---

## Ejemplo

Para un servicio llamado `personas`:

```dockerfile
RUN go build -o build/personas *.go

COPY --from=builder /app/build/personas .

CMD ["./personas","-c","/etc/personas/config.yml","--auto"]
```

---

## Reglas obligatorias

- Usar multi-stage build.
- Usar imagen base `golang:alpine` para build.
- Usar `alpine` para runtime.
- El binario debe llamarse igual que el servicio.
- El path de config debe incluir el nombre del servicio:

```text
/etc/<service_name>/config.yml
```

- No hardcodear nombres de otros servicios.
- No usar nombres genéricos como `app`, `main`, `server`.

---

## Variables importantes

```text
<service_name> = nombre del proyecto
```

Debe ser consistente con:

- nombre del binario
- nombre en Dockerfile
- nombre en Makefile (`image`)
- nombre de carpeta de config (`/etc/...`)

---

## Anti-patterns

No hardcodear otro servicio:

```dockerfile
RUN go build -o build/tareo *.go
```

```dockerfile
CMD ["./tareo","-c","/etc/tareo/config.yml"]
```

No usar nombres genéricos:

```dockerfile
RUN go build -o build/app *.go
```

No usar una sola stage:

```dockerfile
FROM golang:alpine
```

---

## Checklist

- [ ] Usa multi-stage
- [ ] Usa `<service_name>`
- [ ] No hardcodea nombres
- [ ] Usa alpine en runtime
- [ ] Config path correcto