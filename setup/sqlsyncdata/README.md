# SQL Sync Data

Este paquete expone endpoints para sincronizar datos de tablas PostgreSQL entre el servidor y un cliente local. Esta documentacion esta escrita para el cliente que consume los endpoints.

El flujo esta pensado para clientes que mantienen una base local, por ejemplo SQLite, y necesitan recibir cambios del servidor, enviar cambios locales, o ambas cosas segun la configuracion de cada tabla.

## Conceptos

`table_name` es el nombre fisico de la tabla registrada para sincronizacion.

`sync_at` es el checkpoint de sincronizacion en milisegundos Unix. El cliente debe guardar el ultimo checkpoint usado/recibido y enviarlo en la siguiente sincronizacion.

`payload` es la lista de registros que el cliente envia al servidor para insertar o actualizar.

`identifiers` es la lista de columnas que forman la primary key de la tabla.

`read_only` indica que el cliente no debe enviar cambios para esa tabla. Solo puede leer datos desde el servidor.

`write_only` indica que el cliente puede enviar cambios al servidor, pero no recibira registros del servidor para esa tabla.

## Endpoints

### POST `/v1/sync_data/tabla_info`

Devuelve la informacion necesaria para que el cliente prepare sus tablas locales y conozca las reglas de sincronizacion.

Request body:

```json
["tabla_1", "tabla_2"]
```

Response body:

```json
[
  {
    "table_name": "tabla_1",
    "script": "CREATE TABLE IF NOT EXISTS tabla_1(...)",
    "start_sync": 1710000000000,
    "retention_days": 30,
    "read_only": false,
    "write_only": false
  }
]
```

Campos de respuesta:

`table_name`: tabla solicitada.

`script`: SQL que el cliente puede ejecutar localmente para crear la tabla si no existe. Puede incluir indices.

`start_sync`: timestamp inicial recomendado para la primera sincronizacion.

`retention_days`: cantidad de dias hacia atras usada para calcular `start_sync`. Si es `0`, el inicio representa historial completo.

`read_only`: si es `true`, el cliente no debe enviar `payload` para esta tabla.

`write_only`: si es `true`, el cliente puede enviar `payload`, pero el servidor respondera con `payload` vacio.

Uso recomendado:

1. Solicitar la informacion de todas las tablas que el cliente necesita sincronizar.
2. Ejecutar localmente el `script` de cada tabla.
3. Guardar `start_sync`, `read_only`, `write_only` y `retention_days` por tabla.
4. Usar esa metadata para decidir como llamar a `/v1/sync_data/sync`.

### POST `/v1/sync_data/sync`

Sincroniza una tabla. En una sola llamada puede recibir cambios locales del cliente y devolver cambios remotos del servidor.

Request body:

```json
{
  "table_name": "tabla_1",
  "sync_at": 1710000000000,
  "payload": [
    {
      "id": "123",
      "name": "Registro local"
    }
  ]
}
```

Response body:

```json
{
  "identifiers": ["id"],
  "payload": [
    {
      "id": "456",
      "name": "Registro remoto",
      "sync_at": 1710000100000
    }
  ]
}
```

Campos de request:

`table_name`: tabla a sincronizar.

`sync_at`: ultimo checkpoint conocido por el cliente. El servidor devolvera registros con `sync_at` mayor a este valor, salvo que la tabla este configurada como full sync.

`payload`: registros locales que el cliente quiere enviar al servidor. Puede omitirse o enviarse como arreglo vacio si el cliente solo quiere leer.

Campos de respuesta:

`identifiers`: columnas que forman la primary key. El cliente debe usarlas para hacer upsert local.

`payload`: registros que el cliente debe aplicar localmente. Puede venir vacio.

## Modos De Tabla

### Read + Write

Es el modo normal.

El cliente puede enviar `payload` y tambien recibira cambios del servidor.

Comportamiento:

1. El servidor lee cambios remotos segun `sync_at` y el contexto autenticado de la request.
2. El servidor valida el `payload` recibido.
3. El servidor inserta o actualiza el `payload` recibido.
4. El servidor devuelve cambios remotos, excluyendo registros que el cliente acaba de enviar.

### Read Only

El cliente solo puede leer datos del servidor.

Si `read_only` es `true`, el cliente debe llamar a `/v1/sync_data/sync` sin `payload` o con `payload` vacio.

Si envia registros en `payload`, el servidor respondera error y no insertara datos.

### Write Only

El cliente solo envia datos al servidor.

Si `write_only` es `true`, el servidor aceptara `payload`, lo insertara o actualizara, y respondera con `payload` vacio.

Este modo es util para tablas tipo cola, logs, eventos, tracking o datos que suben al servidor pero no deben replicarse de vuelta al cliente.

## Alcance De Datos

El alcance de datos es responsabilidad del servidor.

El cliente no debe enviar ni calcular informacion de alcance para sincronizar.

El servidor determina internamente el contexto autorizado de la request y usa ese contexto para decidir que registros puede leer o escribir el cliente.

Para el cliente, las primary keys son solo identificadores de registros. Debe enviarlas y guardarlas exactamente como vienen en la data de la tabla.

No transformar las primary keys salvo que la propia data de negocio ya las incluya asi.

Si el servidor rechaza un registro por alcance invalido, el cliente debe tratarlo como error de sincronizacion de esa llamada y reintentar o reportar el problema segun su politica local.

## Manejo Local Recomendado

Para cada tabla sincronizada, el cliente deberia guardar:

`table_name`: nombre de la tabla.

`last_sync_at`: ultimo checkpoint usado correctamente.

`read_only`: regla de escritura.

`write_only`: regla de lectura.

`identifiers`: primary keys recibidas desde `/sync`.

Flujo recomendado por tabla:

1. Obtener metadata con `/tabla_info`.
2. Crear o actualizar la tabla local usando `script`.
3. Definir `sync_at` inicial usando `start_sync` si no hay checkpoint local.
4. Si `read_only` es `true`, enviar `payload` vacio.
5. Si `write_only` es `true`, enviar cambios locales y esperar `payload` vacio.
6. Si ambos flags son `false`, enviar cambios locales y aplicar los cambios remotos recibidos.
7. Hacer upsert local usando las columnas de `identifiers`.
8. Avanzar el checkpoint local solo cuando toda la llamada y aplicacion local termine correctamente.

## Errores Esperados

Tabla no registrada:

El servidor responde error si `table_name` no existe o no esta registrada para sincronizacion.

Tabla read-only con payload:

El servidor responde error si se envia `payload` a una tabla marcada como `read_only`.

Registro fuera de alcance:

El servidor puede responder error si un registro del `payload` no pertenece al alcance autorizado de la request. El cliente no debe intentar corregir esto transformando identificadores.

Campo primary key faltante:

El servidor responde error si un registro del `payload` no incluye alguna columna primary key requerida.

Tipo invalido en primary key:

El servidor puede responder error si una primary key requerida tiene un tipo invalido.

## Notas Para IA Y Automatizacion

Antes de llamar `/sync`, siempre consultar `/tabla_info` para conocer `read_only`, `write_only` y el SQL local.

No asumir que todas las tablas son bidireccionales. Usar los flags por tabla.

No enviar `payload` a tablas `read_only`.

No esperar datos remotos desde tablas `write_only`.

Usar `identifiers` para aplicar upsert local, no asumir que la primary key siempre se llama `id`.

Mantener `sync_at` en milisegundos Unix.

Guardar el checkpoint solo despues de persistir localmente todos los cambios recibidos y confirmar que los cambios enviados fueron aceptados.

Si una llamada falla, repetir luego con el mismo checkpoint anterior.

El alcance interno de datos no forma parte del contrato cliente-servidor de estos endpoints. Es informacion interna del backend.
