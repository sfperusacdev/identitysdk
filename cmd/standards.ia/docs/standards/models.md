## Relación con modelos Go

Toda tabla principal que tenga campos de auditoría:

```sql
created_by varchar(100) not null,
created_at timestamp without time zone not null,
updated_by varchar(100) not null,
updated_at timestamp without time zone not null
```

debe tener un modelo Go que embeba:

```go
identitysdk.Model
```

Ejemplo:

```go
package models

import "github.com/sfperusacdev/identitysdk"

type PersonaModel struct {
	Codigo string `json:"codigo" gorm:"primaryKey" chk:"nonil"`
	Nombre string `json:"nombre"`
	Edad   int    `json:"edad"`

	identitysdk.Model
}
```

`identitysdk.Model` define los campos base:

```go
type Model struct {
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedBy string    `json:"updated_by"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

Estos campos son poblados por:

```go
identitysdk.CreateBy(ctx, &entity)
identitysdk.UpdateBy(ctx, &entity)
```