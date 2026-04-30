# Configuración de servicios (ESTÁNDAR)

## Formato

Todos los servicios usan un archivo de configuración en YAML.

## Estructura obligatoria

```yaml
address: "0.0.0.0:7722"

identity: "http://<identity-host>:<port>"
identity_access_token: "<token>"

database:
  host: "<db-host>"
  port: 5432
  db_name: "<db-name>"
  username: "<user>"
  password: "<password>"
  logLevel: "info"
```

---

## Campos

### address

* Dirección donde el servicio expone HTTP.
* Formato: `"host:port"`
* Ejemplo: `"0.0.0.0:7722"`

---

### identity

* URL del servicio de identidad de la empresa.
* Usado para registro y comunicación entre servicios.

---

### identity_access_token

* Token de autenticación para comunicarse con el servicio de identidad.
* Debe ser válido y provisionado externamente.

---

### database

Configuración de conexión a base de datos:

* `host`: dirección del servidor DB
* `port`: puerto (default: 5432)
* `db_name`: nombre de la base
* `username`: usuario
* `password`: contraseña
* `logLevel`: nivel de logging

---

## Reglas obligatorias

* NO hardcodear valores en código.
* SIEMPRE leer desde archivo de configuración.
* SIEMPRE incluir bloque `database`.
* SIEMPRE incluir `identity` y `identity_access_token`.
* El puerto HTTP SIEMPRE viene de `address`.

---

## Seguridad

* `identity_access_token` es sensible.
* NO commitear tokens reales en repositorios.
* Usar valores dummy o variables de entorno en entornos reales.

---

## Anti-patterns

❌ Hardcodear puerto:

```go
http.ListenAndServe(":8080", nil)
```

❌ Configurar DB en código:

```go
db.Connect("localhost", "root", "1234")
```

❌ Omitir identity:

```yaml
# faltante
```

---

## Ejemplo completo

```yaml
address: "0.0.0.0:7722"
identity: "http://192.16.0.15:7771"
identity_access_token: "example-token"

database:
  host: "192.16.0.15"
  port: 5432
  db_name: "cursito_db"
  username: "kevin"
  password: "password"
  logLevel: "info"
```

