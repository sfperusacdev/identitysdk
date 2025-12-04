package turnos

import (
	"fmt"
	"sort"
	"time"
)

func ValidarTurnos(turnos []Turno) ([]Turno, error) {
	if len(turnos) == 0 {
		return nil, fmt.Errorf("lista de turnos vacía")
	}

	type parsed struct {
		Codigo string
		Start  time.Duration
		End    time.Duration
		Raw    Turno
	}

	list := make([]parsed, 0, len(turnos))

	for _, t := range turnos {
		hi, err1 := time.Parse("15:04:05", t.Inicio)
		hf, err2 := time.Parse("15:04:05", t.Fin)
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("formato de hora inválido en turno %s", t.Codigo)
		}

		start := time.Duration(hi.Hour())*time.Hour +
			time.Duration(hi.Minute())*time.Minute +
			time.Duration(hi.Second())*time.Second

		end := time.Duration(hf.Hour())*time.Hour +
			time.Duration(hf.Minute())*time.Minute +
			time.Duration(hf.Second())*time.Second

		list = append(list, parsed{
			Codigo: t.Codigo,
			Start:  start,
			End:    end,
			Raw:    t,
		})
	}

	sort.Slice(list, func(i, j int) bool { return list[i].Start < list[j].Start })

	for i := 0; i < len(list)-1; i++ {
		cur := list[i]
		next := list[i+1]

		curEnd := cur.End
		if curEnd <= cur.Start {
			curEnd += 24 * time.Hour
		}

		if curEnd != next.Start {
			return nil, fmt.Errorf("hay espacio o superposición entre turnos %s y %s", cur.Codigo, next.Codigo)
		}
	}

	total := time.Duration(0)
	for _, p := range list {
		end := p.End
		if end <= p.Start {
			end += 24 * time.Hour
		}
		total += end - p.Start
	}

	if total != 24*time.Hour {
		return nil, fmt.Errorf("los turnos no cubren 24h completas")
	}

	out := make([]Turno, len(list))
	for i := range list {
		out[i] = list[i].Raw
	}

	return out, nil
}
