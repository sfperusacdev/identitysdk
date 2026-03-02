package prorrateo

import (
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

/*
Este algoritmo realiza un prorrateo de valores numéricos basado en el tiempo.

Distribuye el valor de eventos que ocurren en un instante específico entre
un conjunto de entidades que están activas en ese mismo momento. Cada entidad
está definida por un tiempo de inicio y un tiempo de fin. Cuando ocurre un
evento, su valor se divide de forma equitativa entre todas las entidades que
se encuentran activas en ese instante.

El algoritmo utiliza un enfoque de barrido temporal (line sweep):
- los tiempos de inicio y fin de las entidades se convierten en eventos de
  entrada y salida
- los eventos con valor se insertan en la misma línea de tiempo
- la línea de tiempo se ordena cronológicamente
- durante el barrido se mantiene un conjunto de entidades activas
- cuando ocurre un evento con valor, este se reparte entre las entidades activas

El prorrateo se realiza con una precisión fija de 4 decimales. El redondeo
preserva la suma total original: cualquier sobrante generado por el
redondeo se redistribuye en unidades mínimas antes de finalizar el reparto.
*/

type ValueEvent interface {
	Time() time.Time
	Amount() decimal.Decimal
}

type TimeBoundEntity interface {
	StartTime() time.Time
	EndTime() time.Time
}

type Accumulation[E TimeBoundEntity] struct {
	Entity E
	Total  decimal.Decimal
}

func AllocateOverTime[
	E TimeBoundEntity,
	V ValueEvent,
](
	events []V,
	entities []E,
) []Accumulation[E] {

	type timelineItem struct {
		time  time.Time
		kind  int
		ref   *Accumulation[E]
		value decimal.Decimal
	}

	const (
		entityEnter = iota
		entityLeave
		valueOccur
	)

	results := make([]Accumulation[E], len(entities))
	for i, e := range entities {
		results[i] = Accumulation[E]{
			Entity: e,
			Total:  decimal.Zero,
		}
	}

	var timeline []timelineItem

	for i := range results {
		r := &results[i]
		timeline = append(timeline,
			timelineItem{time: r.Entity.StartTime(), kind: entityEnter, ref: r},
			timelineItem{time: r.Entity.EndTime(), kind: entityLeave, ref: r},
		)
	}

	for _, ev := range events {
		timeline = append(timeline,
			timelineItem{
				time:  ev.Time(),
				kind:  valueOccur,
				value: ev.Amount(),
			},
		)
	}

	sort.Slice(timeline, func(i, j int) bool {
		if timeline[i].time.Equal(timeline[j].time) {
			return timeline[i].kind < timeline[j].kind
		}
		return timeline[i].time.Before(timeline[j].time)
	})

	active := make([]*Accumulation[E], 0)

	const scale int32 = 4
	unit := decimal.NewFromFloat(0.0001)

	for _, item := range timeline {
		switch item.kind {
		case entityEnter:
			active = append(active, item.ref)

		case entityLeave:
			for i, a := range active {
				if a == item.ref {
					active = append(active[:i], active[i+1:]...)
					break
				}
			}

		case valueOccur:
			n := len(active)
			if n == 0 {
				continue
			}

			exact := item.value.Div(decimal.NewFromInt(int64(n)))
			base := exact.Truncate(scale)

			baseTotal := base.Mul(decimal.NewFromInt(int64(n)))
			remainder := item.value.Sub(baseTotal)

			extraUnits := remainder.Div(unit).IntPart()

			for i, a := range active {
				if int64(i) < extraUnits {
					a.Total = a.Total.Add(base).Add(unit)
				} else {
					a.Total = a.Total.Add(base)
				}
			}
		}
	}

	return results
}
