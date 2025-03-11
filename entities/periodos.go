package entities

import "time"

type Periodo struct {
	Anio       string
	PlanillaID string
	Items      []PeriodoItem
}

type PeriodoItem struct {
	Item        string
	Mes         string
	FechaInicio time.Time
	FechaFin    time.Time
	Tipo        string
}
