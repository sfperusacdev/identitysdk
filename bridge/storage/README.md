# Bridge Storage

`bridge/storage` crea `FileStorer` usando variables leidas desde `bridge/variables`.

## Uso

```go
storer, err := bridge.Storage.Create(ctx, "bucket/path")
```

Para variables globales:

```go
storer, err := bridge.Storage.CreateGlobal(ctx, "bucket/path")
```

Para usar un lector de variables custom:

```go
storer, err := bridge.Storage.CreateWith(ctx, "bucket/path", readVariable)
```

## FileStorer

La interfaz principal esta compuesta por interfaces pequenas:

```go
type FileStorer interface {
	Reader
	Writer
	Remover
	Lister
}
```

Metodos base:

```go
Read(ctx, name)
Write(ctx, name, data)
WriteFrom(ctx, name, reader)
Remove(ctx, name)
List(ctx, prefix)
```

Operaciones derivadas disponibles como helpers del paquete:

```go
storage.Replace(ctx, storer, name, data)
storage.WriteBatch(ctx, storer, files)
storage.WriteFromBatch(ctx, storer, files)
```

## Driver

El driver se define con:

```text
STORAGE_DRIVER
```

Valores soportados:

```text
minio
aws-s3
local
```

Si `STORAGE_DRIVER` no existe o esta vacio, se usa `minio`.

## MinIO

Variables:

```text
STORAGE_MINIO_ACCESS_KEY_ID
STORAGE_MINIO_SECRET_ACCESS_KEY
STORAGE_MINIO_ENDPOINT
STORAGE_MINIO_REGION
```

`STORAGE_MINIO_ENDPOINT` soporta valores con o sin esquema:

```text
http://localhost:9000
https://minio.example.com
localhost:9000
```

Si `STORAGE_MINIO_REGION` no existe o esta vacio, se usa `us-east-1`.

Fallback temporal:

```text
MINIO_ACCESS_KEY_ID
MINIO_SECRET_ACCESS_KEY
MINIO_ENDPOINT
MINIO_REGION
```

## AWS S3

Variables:

```text
STORAGE_AWS_S3_ACCESS_KEY_ID
STORAGE_AWS_S3_SECRET_ACCESS_KEY
STORAGE_AWS_S3_BASE_URL
STORAGE_AWS_S3_REGION
```

`STORAGE_AWS_S3_BASE_URL` es opcional.

Si `STORAGE_AWS_S3_REGION` no existe o esta vacio, se usa `us-east-1`.

Fallback temporal:

```text
AWS_S3_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID
AWS_S3_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY
AWS_S3_BASE_URL
S3_SDK_STORAGE_BASE_URL
AWS_S3_REGION
AWS_REGION
```

## Local

Variables:

```text
STORAGE_LOCAL_BASE_PATH
```

Fallback temporal:

```text
LOCAL_STORAGE_BASE_PATH
```

El path final se construye con:

```text
STORAGE_LOCAL_BASE_PATH + bucket
```

## Compatibilidad

`services.ExternalBridgeService.CreateNewFileStorer` sigue disponible, pero esta marcado como deprecated.

Usar en su lugar:

```go
bridge.Storage.Create(ctx, bucket)
bridge.Storage.CreateGlobal(ctx, bucket)
```
