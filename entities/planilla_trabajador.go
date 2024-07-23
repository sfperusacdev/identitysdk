package entities

import "time"

type TrabajadorPlanilla struct {
	Codigo               string    `json:"codigo"`
	PlanillaCodigo       string    `json:"planilla_codigo"`
	Inactivo             bool      `json:"inactivo"`
	FechaIngreso         string    `json:"fecha_ingreso"`
	FueLiquidado         bool      `json:"fue_liquidado"`
	GrupoProvicionCodigo string    `json:"grupo_provicion_codigo"`
	CreatedAt            time.Time `json:"created_at"`
}
