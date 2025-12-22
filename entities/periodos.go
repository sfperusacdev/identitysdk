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

func (p PeriodoItem) IsZero() bool {
	return p.Item == "" &&
		p.Mes == "" &&
		p.FechaInicio.IsZero() &&
		p.FechaFin.IsZero() &&
		p.Tipo == ""
}

type Periodos []Periodo

var PeriodosData Periodos

func (ps Periodos) BuscarPeriodoDia(
	planilla string,
	fecha time.Time,
) PeriodoItem {
	for _, p := range ps {
		if p.PlanillaID == planilla {
			for _, itm := range p.Items {
				if (itm.FechaInicio.Equal(fecha) || itm.FechaInicio.Before(fecha)) &&
					(itm.FechaFin.Equal(fecha) || itm.FechaFin.After(fecha)) {
					return itm
				}
			}
		}
	}
	return PeriodoItem{}
}
