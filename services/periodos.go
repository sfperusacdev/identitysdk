package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/entities"
	"github.com/sfperusacdev/identitysdk/xreq"
)

func (s *ExternalBridgeService) Periodos(ctx context.Context) ([]entities.Periodo, error) {
	company, token := s.readCompanyAndToken(ctx)
	baseURL, err := identitysdk.GetAsistenciaServiceURL(ctx, company)
	if err != nil {
		slog.Error("error retrieving service URL", "error", err)
		return nil, err
	}

	var response struct {
		Message string `json:"message"`
		Data    []struct {
			PlanillaID  string    `json:"planilla_id"`
			Anio        string    `json:"anio"`
			Mes         string    `json:"mes"`
			Item        string    `json:"item"`
			FechaInicio time.Time `json:"fecha_inicio"`
			FechaFin    time.Time `json:"fecha_fin"`
			Tipo        string    `json:"tipo"`
		} `json:"data"`
	}

	endpoint := "v1/api/listar-periodos"

	if err := xreq.MakeRequest(ctx,
		baseURL, endpoint,
		xreq.WithAuthorization(token),
		xreq.WithUnmarshalResponseInto(&response),
	); err != nil {
		return nil, err
	}

	periodMap := make(map[string]*entities.Periodo)
	orderedKeys := make([]string, 0)

	for _, d := range response.Data {
		key := d.Anio + "_" + d.PlanillaID
		if _, exists := periodMap[key]; !exists {
			periodMap[key] = &entities.Periodo{
				Anio:       d.Anio,
				PlanillaID: d.PlanillaID,
				Items:      make([]entities.PeriodoItem, 0),
			}
			orderedKeys = append(orderedKeys, key)
		}
		periodMap[key].Items = append(periodMap[key].Items, entities.PeriodoItem{
			Item:        d.Item,
			Mes:         d.Mes,
			FechaInicio: d.FechaInicio,
			FechaFin:    d.FechaFin,
			Tipo:        d.Tipo,
		})
	}

	result := make([]entities.Periodo, 0, len(orderedKeys))
	for _, k := range orderedKeys {
		result = append(result, *periodMap[k])
	}

	return result, nil
}
