## IdentitySDK

libreria de apoyo para validar un token contra el servicio de identity

## Logs y OpenTelemetry

`setup.NewService` incluye telemetría configurable desde el archivo actual del
servicio. Los logs `slog` siguen saliendo por stdout, con o sin exportación OTLP.
Sin sección `telemetry`, la exportación permanece desactivada.

Ver [configuración, defaults y cobertura](telemetry/README.md).
