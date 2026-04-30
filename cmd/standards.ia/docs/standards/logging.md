# Logging

## Objetivo

Estandarizar el uso de logs en el sistema.

La librería estándar es:

```go
"log/slog"
```

---

## Filosofía

El logging no es obligatorio en todas partes.

Los logs deben usarse **solo cuando aportan valor**.

---

## Cuándo loggear

Se debe loggear cuando:

- ocurre un error relevante
- hay interacción con servicios externos
- hay operaciones críticas del sistema
- se necesita trazabilidad

Ejemplos:

```go
slog.Error("Error al consumir servicio externo", "error", err)
```

```go
slog.Error("Error al procesar pago", "codigo", codigo, "error", err)
```

---

## Cuándo NO loggear

No loggear:

- operaciones triviales
- flujos normales sin errores
- cada paso del código
- logs redundantes

Ejemplo incorrecto:

```go
slog.Info("Entrando a función")
slog.Info("Validando payload")
slog.Info("Llamando usecase")
```

---

## Errores y logging

El sistema usa:

```go
github.com/user0608/goones/errs
```

El paquete `errs` puede manejar logging internamente.

---

## Regla importante

No duplicar logs.

Incorrecto:

```go
slog.Error("Error DB", "error", err)
return errs.Pgf(err)
```

Correcto:

```go
return errs.Pgf(err)
```

---

## Dónde loggear

### Handler

Solo si aporta valor adicional:

- errores inesperados
- debugging puntual

---

### Usecase

Loggear solo si:

- hay integración externa
- hay lógica crítica

---

### Repository

Generalmente NO loggear.

Los errores se manejan con:

```go
return errs.Pgf(err)
```

---

## Logs en servicios externos

Aquí sí es recomendable loggear:

```go
resp, err := client.Do(req)
if err != nil {
	slog.Error("Error llamando servicio externo", "url", url, "error", err)
	return err
}
```

---

## Formato

Usar siempre formato estructurado:

```go
slog.Error("mensaje", "key", value)
```

No concatenar strings:

```go
slog.Error("error: " + err.Error()) ❌
```

---

## Datos sensibles

Nunca loggear:

- passwords
- tokens
- identity_access_token
- datos personales sensibles

---

## Reglas obligatorias

- Usar `slog`
- Loggear errores relevantes
- No loggear todo indiscriminadamente
- No duplicar logs con `errs`
- Usar logs estructurados
- No exponer datos sensibles

---

## Anti-patterns

No usar:

```go
fmt.Println(err)
```

```go
log.Println(err)
```

No loggear todo:

```go
slog.Info("paso 1")
slog.Info("paso 2")
slog.Info("paso 3")
```

No duplicar logs:

```go
slog.Error("error", "error", err)
return errs.Pgf(err)
```

---

## Checklist

- [ ] Usa slog
- [ ] Logs solo donde aportan valor
- [ ] No hay logs redundantes
- [ ] No se duplican logs con errs
- [ ] Logs estructurados
- [ ] No expone datos sensibles