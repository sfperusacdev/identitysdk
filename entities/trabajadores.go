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
