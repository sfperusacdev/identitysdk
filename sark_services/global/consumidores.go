package global

import (
	"context"

	"github.com/sfperusacdev/identitysdk/xreq"
)

type ConsumidorTipo struct {
	Codigo      string `json:"codigo"`
	Descripcion string `json:"descripcion"`
}

type Consumidor struct {
	Codigo            string          `json:"codigo"`
	Descripcion       string          `json:"descripcion"`
	Padre             *string         `json:"padre"`
	Tipo              *ConsumidorTipo `json:"tipo"`
	ReferenciaExterna string          `json:"referencia_externa"`
	Hierarchy         []string        `json:"hierarchy"`
}

func (s *GlobalService) Consumidores(ctx context.Context) ([]Consumidor, error) {
	apiurl, err := s.env.GetGeneralServiceURL(ctx, s.env.Empresa(ctx))
	if err != nil {
		return nil, err
	}

	var apiResponse struct {
		Data []Consumidor `json:"data"`
	}

	if err := xreq.MakeRequest(ctx, apiurl,
		"/api/v1/consumidores",
		xreq.WithQueryParam("sucursal", s.env.Sucursal(ctx)),
		xreq.WithUnmarshalResponseInto(&apiResponse),
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
	); err != nil {
		return nil, err
	}
	return apiResponse.Data, nil
}
