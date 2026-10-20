# Telemetría de IdentitySDK

`setup.NewService` configura automáticamente logs y trazas. No es
necesario agregar otro módulo Fx ni cambiar los constructores de los servicios.
La configuración se lee desde el mismo archivo Viper del servicio, incluso con
un `WithConfigProvider` personalizado; no se amplían sus interfaces.

## Organización

Un solo paquete `telemetry` reúne configuración, `TelemetryService`, integración
Fx e instrumentación de trazas y logs. Su constructor es `NewTelemetryService`,
siguiendo el patrón de servicios del proyecto.

```text
telemetry/
  config.go    # Configuración y validación
  service.go   # Inicialización, estado global y apagado
  signals.go   # Recursos, providers, exportadores y privacidad de trazas
  logger.go    # Handler slog, correlación, contexto y diagnósticos
  http.go      # Cliente, middleware y access log
  grpc.go      # Stats handlers de cliente y servidor
  database.go  # Instrumentación GORM
  module.go    # Ciclo de vida Fx
```

`service.go` instala/restaura los defaults de trazas y logging, sin modificar el
provider global de métricas. Las pruebas unitarias acompañan a sus
implementaciones; las pruebas extremo a extremo de la API están separadas por
componente y comparten el Collector simulado de `collector_test.go`.

## Activación

El archivo [config.example.yml](config.example.yml) activa telemetría con la
configuración mínima. Las opciones adicionales están comentadas con sus valores
por defecto. Copiar su sección al archivo actual del servicio y ajustar el endpoint.
Los logs continúan saliendo por stdout.

Agregar al archivo de configuración actual:

```yaml
telemetry:
  enabled: true
  environment: production
  otlp:
    endpoint: "http://otel-collector:4318"
```

Sin esta sección, o con `enabled: false`, no se crean exportadores y los logs
siguen saliendo por **stdout**. Con `enabled: true` stdout permanece habilitado:
OpenTelemetry agrega un destino, nunca reemplaza la consola.

| Campo | Default | Regla |
| --- | --- | --- |
| `enabled` | `false` | Interruptor general de telemetría. |
| `environment` | vacío | Recomendado; no se supone `production`. |
| `otlp.endpoint` | vacío | Obligatorio si se exporta alguna señal. URL HTTP(S) sin credenciales ni query. |
| `otlp.protocol` | `http/protobuf` | También admite `grpc`. |
| `traces.enabled` | `true` | Aplicado cuando el interruptor general está activo. |
| `traces.sample_ratio` | `0.1` | De 0 a 1; respeta el muestreo del padre. `0` explícito se conserva. |
| `logs.level` | `info` | `debug`, `info`, `warn`, `error`. |
| `logs.format` | `json` | `json` o `text` para stdout. |
| `logs.export_otlp` | `true` | Solo controla exportación de logs, no stdout. |

Las opciones de logs de esta sección se aplican cuando `enabled: true`. Sin
telemetría se utiliza stdout en texto con nivel info.

Para desarrollo, con Collector local:

```yaml
telemetry:
  enabled: true
  environment: development
  otlp:
    endpoint: "http://localhost:4318"
  logs:
    level: debug
    format: text
    export_otlp: false
```

Este ejemplo envía trazas, pero conserva los logs únicamente en consola.
Sin Collector, omitir `telemetry` o usar `enabled: false`.

Para OTLP/gRPC usar `protocol: grpc` y, habitualmente,
`endpoint: "http://otel-collector:4317"`. Para TLS, usar `https://...` y los
certificados del Collector. El endpoint HTTP es base: se agregan `/v1/traces`
y `/v1/logs`, admitiendo un prefijo de proxy.

El archivo gobierna endpoint, protocolo y switches. Los exportadores oficiales
permiten configurar autenticación y certificados mediante sus variables estándar,
por ejemplo `OTEL_EXPORTER_OTLP_HEADERS` y `OTEL_EXPORTER_OTLP_CERTIFICATE` (también
sus variantes por señal). Guardar esos valores en secretos de despliegue, no en
el YAML. `OTEL_RESOURCE_ATTRIBUTES` permite atributos adicionales; nombre y versión
del servicio se obtienen del setup, y cada arranque genera un `service.instance.id`.

## Logs existentes

No cambia la API:

```go
slog.Info("operación completada", "codigo", codigo)
slog.Error("falló la operación", "error", err)
```

Para correlacionar un log con una traza, usar el contexto recibido:

```go
slog.ErrorContext(ctx, "falló la operación", "error", err)
```

Los handlers agregan `trace_id` y `span_id` a stdout y el bridge `otelslog`
establece la correlación nativa en OTLP. Los logs sin contexto se exportan igual,
pero no se pueden asociar automáticamente a una petición. Un logger creado por
la aplicación con `slog.New` y un handler propio no usa el handler global del SDK.

Los eventos de Fx y el access log de Echo usan `slog`; el access log contiene
plantillas de rutas, estado y duración, sin query strings. Los valores de empresa
y sucursal del contexto autenticado se agregan a logs con contexto y trazas.
No se usan para autenticación ni autorización.

## Cobertura

- HTTP entrante de `httpapi` y saliente de `xreq`, resolución de servicios,
  validación en Identity y los helpers HTTP comunes. No se modifica globalmente
  `http.DefaultClient` ni su transporte.
- Servidor y cliente gRPC del SDK, unary y streaming, conservando autenticación.
- GORM de la conexión devuelta al arrancar por `StorageManager`, incluyendo
  proveedores personalizados: spans de operaciones, duración y estado de error.
- Propagación W3C `traceparent`/`tracestate`; no se propaga baggage del usuario.
- `CloneContext` y `FireAndForget` conservan la identidad inmutable de la traza,
  sin conservar cancelación, deadline, el span vivo ni valores arbitrarios.

No se recopilan ni exportan métricas mediante esta integración. HTTP y gRPC usan
providers de métricas no-op y GORM tiene sus métricas desactivadas. No se modifica
el provider global de métricas que otra aplicación pudiera configurar.
El endpoint `/metrics` JSON existente del SDK se mantiene independiente y sin cambios.

Las conexiones HTTP/gRPC/DB creadas fuera de los componentes comunes necesitan
instrumentación adicional. La conexión inicial y migraciones se ejecutan antes
de Fx: sus logs ya usan el logger configurado, pero las consultas ejecutadas
directamente sobre `database/sql` no pasan por los callbacks de GORM. RabbitMQ,
storage, spans de negocio y despliegue del Collector quedan fuera de esta etapa.

## Seguridad y operación

No se capturan cuerpos ni headers de autenticación. El exportador de trazas
elimina query strings/credenciales de URLs, SQL completo, mensajes de excepciones
y descripciones de error potencialmente sensibles. Conserva estado de error,
tipo de excepción, operación y tiempos. **Esto no sanea mensajes de logs escritos
por la aplicación:** nunca registrar JWT, claves, contraseñas ni datos personales.

La inicialización no espera a que el Collector esté disponible. La exportación
es asíncrona, con colas acotadas y timeouts: una caída del Collector no bloquea
las peticiones; puede perderse telemetría. Los avisos de exportación fallida van
solo a stdout, limitados a uno por minuto, para evitar recursión y ruido.

Fx drena servidores antes de cerrar los providers. Se intenta vaciar cada señal
durante el apagado, respetando el deadline; también se liberan recursos si falla
el arranque. `TelemetryService` es de alcance de proceso, igual que
`slog.SetDefault` y los providers globales de OpenTelemetry; no iniciar varias
instancias concurrentes en el mismo proceso.

Si logs se envían por OTLP, evitar recolectar también stdout hacia el mismo
backend para no duplicarlos. OpenTelemetry no es el almacenamiento: desplegar un
Collector y configurar sus destinos por separado.

## Pruebas

```sh
go test -race ./telemetry/... ./xreq ./httpapi ./grpc/... .
```

Las pruebas usan receptores OTLP HTTP/gRPC simulados, Echo, gRPC en memoria y SQL
en memoria. Verifican señales, correlación, contexto, privacidad, muestreo,
stdout, autenticación, apagado y Collector no disponible, sin backend externo.
