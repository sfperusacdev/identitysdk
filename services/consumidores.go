package services

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/xreq"
)

type TareoConsumidorTipoDto struct {
	Codigo      string `json:"codigo"`
	Descripcion string `json:"descripcion"`
}

type TareoConsumidorDto struct {
	Codigo            string                  `json:"codigo"`
	Descripcion       string                  `json:"descripcion"`
	Padre             *string                 `json:"padre"`
	Tipo              *TareoConsumidorTipoDto `json:"tipo"`
	ReferenciaExterna string                  `json:"referencia_externa"`
	Hierarchy         []string                `json:"hierarchy"`
}

func (s *ExternalBridgeService) GetConsumidores(ctx context.Context, sucursal string) ([]TareoConsumidorDto, error) {
	company, token := s.readCompanyAndToken(ctx)

	baseurl, err := identitysdk.GetTareoServiceURL(ctx, company)
	if err != nil {
		slog.Error("error trying to retrieve `tareo` service url", "error", err)
		return nil, err
	}

	var apiresponse struct {
		Type string               `json:"type"`
		Data []TareoConsumidorDto `json:"data"`
	}

	queryParams := make(url.Values)
	queryParams.Set("sucursal", sucursal)

	if err := xreq.MakeRequest(
		ctx,
		baseurl,
		"/api/v1/consumidores",
		xreq.WithAuthorization(token),
		xreq.WithQueryParams(queryParams),
		xreq.WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return nil, err
	}

	return apiresponse.Data, nil
}
