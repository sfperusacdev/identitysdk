package entities

import "time"

type ResumenTrabajadorDto struct {
	Codigo          string
	Nombres         string
	ApellidoMaterno string
	ApellidoPaterno string
	Documento       ResumenTrabajadorDocumentoIdentidadDto
	Cargo           ResumenTrabajadorCargoDto
	PlanillaCodigo  string
	FechaIngreso    string
	Email           string
	Telefono        string
	Sexo            string
	DeletedAt       *time.Time
}

type ResumenTrabajadorCargoDto struct {
	Codigo      string
	Descripcion string
}

type ResumenTrabajadorDocumentoIdentidadDto struct {
	Tipo   string
	Numero string
}

type TrabajadorDto struct {
	ApellidoMaterno        string     `json:"apellido_materno"`
	ApellidoPaterno        string     `json:"apellido_paterno"`
	Cargo                  string     `json:"cargo"`
	CargoCodigo            string     `json:"cargo_codigo"`
	Codigo                 string     `json:"codigo"`
	Dni                    string     `json:"dni"`
	DocumentoIdentidad     string     `json:"documento_identidad"`
	Edad                   string     `json:"edad"`
	Email                  string     `json:"email"`
	Estado                 bool       `json:"estado"`
	FechaIngreso           string     `json:"fecha_ingreso"`
	ImageLocation          string     `json:"image_location"`
	Nombres                string     `json:"nombres"`
	PlanillaCodigo         string     `json:"planilla_codigo"`
	Sexo                   string     `json:"sexo"`
	Telefono               string     `json:"telefono"`
	TipoDocumentoIdentidad string     `json:"tipo_documento_identidad"`
	DeletedAt              *time.Time `json:"deleted_at"`
}
