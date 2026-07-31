# datasync

El paquete `datasync` compara una coleccion externa con una coleccion local y
ejecuta las operaciones necesarias para sincronizarlas.

La comparacion se define mediante dos funciones:

- `Equals(ext, loc)`: determina si un elemento externo corresponde a uno local.
- `Map(ext)`: convierte un elemento externo al tipo local.

`Equals` y `Map` son obligatorias. Las funciones de escritura son opcionales.

## Regla de comparacion

La decision se toma usando `Equals`, no comparando automaticamente todos los
campos. `Equals` normalmente compara una clave estable, por ejemplo un ID:

```go
Equals: func(ext External, loc Local) bool {
	return ext.ID == loc.ID
},
```

Para cada elemento externo, el paquete busca un elemento local que cumpla
`Equals(ext, loc)`. La coincidencia debe ser uno a uno: un elemento local que
ya fue usado no puede volver a coincidir con otro elemento externo.

El contenido de `Map(ext)` se usa como el nuevo valor local, pero no decide si
existe una coincidencia. Por ejemplo, si el ID coincide pero cambia `Name`, se
ejecuta `Update`; el paquete no determina por si mismo si el cambio es
relevante.

## Cuando se ejecuta cada operacion

| Situacion | Callback | Valores recibidos | Contador |
| --- | --- | --- | --- |
| Existe una coincidencia local | `Update` / `UpdateBatch` | Local anterior y valor nuevo | `Updated` |
| Existe una coincidencia local, pero no hay callback de update | Ninguno | - | `Unchanged` |
| No existe coincidencia local | `Insert` / `InsertBatch` | Valor nuevo mapeado | `Inserted` |
| Un elemento local no fue usado por ningun externo | `Delete` / `DeleteBatch` | Valor local anterior | `Deleted` |
| La operacion no tiene callback | Ninguno | - | No se cuenta |

En otras palabras:

```text
por cada externo:
    nuevo = Map(externo)

    si existe un local no usado donde Equals(externo, local):
        si existe Update: ejecutar Update(local, nuevo)
        si no: contar Unchanged
    si no:
        si existe Insert: ejecutar Insert(nuevo)

por cada local no usado:
    si existe Delete: ejecutar Delete(local)
```

El callback `Update` se ejecuta para toda coincidencia, incluso si los campos
del valor nuevo son iguales a los del valor anterior. Si se quieren evitar
updates innecesarios, esa decision debe hacerse dentro de `Equals` o dentro del
callback de update.

Ejemplo de decision:

```text
Externos: [ID 1, ID 2]
Locales:   [ID 1, ID 3]

ID 1 coincide -> Update(ID 1 local, ID 1 nuevo)
ID 2 no coincide -> Insert(ID 2 nuevo)
ID 3 no fue usado -> Delete(ID 3 local)
```

## Resultado

`Sync` y `SyncBatch` devuelven un `SyncResult` con estos contadores:

- `Inserted`: elementos externos sin coincidencia que fueron insertados.
- `Updated`: elementos con coincidencia cuyo callback de actualizacion se ejecuto.
- `Deleted`: elementos locales sin coincidencia que fueron eliminados.
- `Unchanged`: elementos con coincidencia cuyo callback de actualizacion no esta configurado.

Si una operacion no tiene callback, se omite y no se cuenta como insertada,
actualizada o eliminada. Un elemento coincidente sin callback de actualizacion
se cuenta como `Unchanged`.

## Sincronizacion individual

Use `Sync` cuando cada operacion debe ejecutarse por separado:

```go
strategy := datasync.SyncStrategy[External, Local]{
	Equals: func(ext External, loc Local) bool {
		return ext.ID == loc.ID
	},
	Map: func(ext External) Local {
		return Local{ID: ext.ID, Name: ext.Name}
	},
	Insert: func(ctx context.Context, new Local) error {
		return repository.Insert(ctx, new)
	},
	Update: func(ctx context.Context, old Local, new Local) error {
		return repository.Update(ctx, old, new)
	},
	Delete: func(ctx context.Context, old Local) error {
		return repository.Delete(ctx, old)
	},
}

result, err := datasync.Sync(ctx, external, local, strategy)
```

El orden de ejecucion es:

1. Recorre los elementos externos.
2. Actualiza las coincidencias o inserta los elementos nuevos.
3. Elimina los elementos locales que no tuvieron coincidencia.

El orden es significativo: una llamada a `Sync` puede ejecutar varios inserts o
updates antes de intentar los deletes. Si una operacion falla, las operaciones
anteriores no se revierten.

Las funciones `Update` reciben tanto el valor local anterior como el nuevo
valor mapeado. `Delete` recibe el valor local anterior.

## Sincronizacion por lotes

Use `SyncBatch` para acumular cada tipo de operacion y ejecutar un callback por
lote:

```go
strategy := datasync.SyncBatchStrategy[External, Local]{
	Equals: func(ext External, loc Local) bool {
		return ext.ID == loc.ID
	},
	Map: func(ext External) Local {
		return Local{ID: ext.ID, Name: ext.Name}
	},
	InsertBatch: func(ctx context.Context, values []Local) error {
		return repository.InsertBatch(ctx, values)
	},
	UpdateBatch: func(ctx context.Context, oldValues []Local, newValues []Local) error {
		return repository.UpdateBatch(ctx, oldValues, newValues)
	},
	DeleteBatch: func(ctx context.Context, values []Local) error {
		return repository.DeleteBatch(ctx, values)
	},
}

result, err := datasync.SyncBatch(ctx, external, local, strategy)
```

Los lotes se ejecutan en este orden:

1. `InsertBatch` con los elementos nuevos.
2. `UpdateBatch` con dos slices paralelos: valores anteriores y valores nuevos.
3. `DeleteBatch` con los elementos locales que sobran.

Cada callback solo se ejecuta cuando su lote no esta vacio y el callback esta
configurado.

`UpdateBatch` recibe dos slices con la misma longitud y el mismo orden. El
elemento `oldValues[i]` corresponde a `newValues[i]`.

Por ejemplo, para:

```text
Externos: [ID 1, ID 2]
Locales:   [ID 1, ID 3]
```

los lotes contienen:

```text
InsertBatch: [ID 2 nuevo]
UpdateBatch: old=[ID 1 local], new=[ID 1 nuevo]
DeleteBatch: [ID 3 local]
```

## Coincidencias y duplicados

Cada elemento local solo puede ser usado una vez. Si varios elementos externos
coinciden con el mismo elemento local, el primero usa esa coincidencia y los
siguientes se consideran nuevos.

Por tanto, la coleccion externa debe contener identificadores unicos cuando el
callback de insercion no pueda aceptar duplicados.

## Errores y transacciones

Si `Equals` o `Map` no estan configuradas, la funcion devuelve un error sin
ejecutar operaciones.

Si un callback devuelve un error, la sincronizacion se detiene y devuelve el
error junto con el resultado acumulado hasta ese momento. El paquete no crea
transacciones ni hace rollback; la atomicidad debe gestionarse dentro de los
callbacks o en la capa de persistencia.

Los callbacks reciben el `context.Context` proporcionado por el llamador y son
responsables de respetar su cancelacion.

## Complejidad

El emparejamiento actual compara cada elemento externo con los elementos
locales, por lo que su complejidad es `O(len(external) * len(local))`. `Sync`
usa memoria adicional `O(len(local))`; `SyncBatch` tambien almacena los lotes
pendientes.
