package services

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/xreq"
)

type TareoLaborDto struct {
	Codigo            string `json:"codigo"`
	Descripcion       string `json:"descripcion"`
	ActividadCodigo   string `json:"actividad_codigo"`
	ReferenciaExterna string `json:"referencia_externa"`
}

type TareoActividadDto struct {
	Codigo            string          `json:"codigo"`
	Descripcion       string          `json:"descripcion"`
	ReferenciaExterna string          `json:"referencia_externa"`
	EsRendimiento     bool            `json:"es_rendimiento"`
	Labores           []TareoLaborDto `json:"labores"`
}

func (s *ExternalBridgeService) GetActividades(ctx context.Context, sucursal string) ([]TareoActividadDto, error) {
	company, token := s.readCompanyAndToken(ctx)

	baseurl, err := identitysdk.GetTareoServiceURL(ctx, company)
	if err != nil {
		slog.Error("error trying to retrieve `tareo` service url", "error", err)
		return nil, err
	}

	var apiresponse struct {
		Type string              `json:"type"`
		Data []TareoActividadDto `json:"data"`
	}

	queryParams := make(url.Values)
	queryParams.Set("sucursal", sucursal)

	if err := xreq.MakeRequest(
		ctx,
		baseurl,
		"/v1/actividades",
		xreq.WithAuthorization(token),
		xreq.WithQueryParams(queryParams),
		xreq.WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return nil, err
	}

	return apiresponse.Data, nil
}
