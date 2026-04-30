# Manejo de errores

## Objetivo

Estandarizar cómo se crean, envuelven, propagan y responden errores en todo el proyecto.

El paquete estándar es:

```go
"github.com/user0608/goones/errs"
```

---

## Flujo estándar

```text
repository → errs.Pgf
usecase    → errs.*
handler    → answer.Err / answer.JsonErr
```

---

## Tipo base de error

`errs` usa un error interno con:

```go
type Err struct {
	httpCode int
	wrapped  error
	message  string
}
```

Este error permite transportar:

- código HTTP
- mensaje para cliente
- error interno envuelto

---

## Crear errores de negocio

Usar errores directos cuando solo se necesita un mensaje:

```go
return errs.BadRequestDirect("codigo requerido")
```

```go
return errs.NotFoundDirect("persona no encontrada")
```

```go
return errs.UnauthorizedDirect("usuario no autorizado")
```

```go
return errs.ForbiddenDirect("no tiene permisos para esta accion")
```

---

## Crear errores con formato

Cuando el mensaje requiere interpolación:

```go
return errs.BadRequestf("codigo %s inválido", codigo)
```

```go
return errs.NotFoundf("persona %s no encontrada", codigo)
```

```go
return errs.InternalErrorf("error procesando operación %s", operacion)
```

---

## Envolver errores existentes

Cuando existe un error original y se quiere cambiar el mensaje visible:

```go
return errs.BadRequestError(err, "request inválido")
```

```go
return errs.NotFoundError(err, "registro no encontrado")
```

```go
return errs.InternalError(err, "error interno")
```

También puede usarse:

```go
return errs.NewWithMessage(err, "mensaje personalizado")
```

---

## Repository

Los errores de PostgreSQL/GORM siempre deben envolverse con:

```go
errs.Pgf(rs.Error)
```

Ejemplo:

```go
rs := tx.Raw(qry, prefix).Scan(&personas)
if rs.Error != nil {
	return nil, errs.Pgf(rs.Error)
}
```

`errs.Pgf` traduce errores PostgreSQL a mensajes y códigos HTTP estándar; también maneja casos como `record not found` y códigos PG específicos. :contentReference[oaicite:0]{index=0}

---

## Usecase

El usecase crea errores de negocio con `errs`.

```go
if persona.Nombre == "" {
	return errs.BadRequestDirect("no se encontro el nombre")
}
```

Los errores de repository o services se propagan:

```go
if err := uc.repository.CrearPersona(ctx, persona); err != nil {
	return err
}
```

---

## Handler

El handler convierte errores a respuesta HTTP usando `answer`.

```go
return answer.Err(c, err)
```

Para JSON inválido:

```go
return answer.JsonErr(c)
```

Para validación con `kcheck`:

```go
return answer.Err(c, errs.BadRequestDirect(err.Error()))
```

---

## Helpers de inspección

Para saber si un error pertenece al estándar:

```go
errs.IsErr(err)
```

Para verificar BadRequest:

```go
errs.IsBadRequest(err)
```

Para verificar InternalError:

```go
errs.IsInternalError(err)
```

Para obtener resumen:

```go
summary := errs.ToSummary(err)
```

Para buscar texto en el mensaje:

```go
errs.ContainsMessage(err, "mensaje")
```

---

## Reglas obligatorias

- Repository usa `errs.Pgf` para errores de base de datos.
- Usecase usa `errs` para errores de negocio.
- Handler responde errores con `answer.Err`.
- No responder errores desde usecase.
- No usar `c.JSON` para errores.
- No usar `http.Error`.
- No retornar errores DB crudos.
- No usar `errors.New` si existe un helper de `errs`.
- Propagar errores sin convertirlos innecesariamente.

---

## Anti-patterns

No retornar error DB directo:

```go
return rs.Error
```

No crear errores genéricos:

```go
return errors.New("codigo requerido")
```

No crear errores HTTP en usecase:

```go
return echo.NewHTTPError(400, "error")
```

No responder desde usecase:

```go
return answer.Err(c, err)
```

No responder manualmente en handler:

```go
return c.JSON(400, map[string]string{"error": err.Error()})
```

No ocultar errores PostgreSQL sin `errs.Pgf`:

```go
return errs.InternalErrorDirect("error en base de datos")
```

---

## Checklist

- [ ] Repository usa `errs.Pgf`.
- [ ] Usecase usa `errs.*`.
- [ ] Handler usa `answer.Err`.
- [ ] JSON inválido usa `answer.JsonErr`.
- [ ] Errores de validación usan `errs.BadRequestDirect`.
- [ ] No se usa `c.JSON` para errores.
- [ ] No se usa `http.Error`.
- [ ] No se retorna `rs.Error` directamente.
- [ ] No se usan errores HTTP fuera del handler.