package services

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/xreq"
)

type TareoVariedadDto struct {
	Codigo            string `json:"codigo"`
	CultivoCodigo     string `json:"cultivo_codigo"`
	Descripcion       string `json:"descripcion"`
	ReferenciaExterna string `json:"referencia_externa"`
}

type TareoCultivoDto struct {
	Codigo            string             `json:"codigo"`
	Descripcion       string             `json:"descripcion"`
	ReferenciaExterna string             `json:"referencia_externa"`
	Variedades        []TareoVariedadDto `json:"variedades"`
}

func (s *ExternalBridgeService) GetCultivos(ctx context.Context, sucursal string) ([]TareoCultivoDto, error) {
	company, token := s.readCompanyAndToken(ctx)

	baseurl, err := identitysdk.GetTareoServiceURL(ctx, company)
	if err != nil {
		slog.Error("error trying to retrieve `tareo` service url", "error", err)
		return nil, err
	}

	var apiresponse struct {
		Type string            `json:"type"`
		Data []TareoCultivoDto `json:"data"`
	}

	queryParams := make(url.Values)
	queryParams.Set("sucursal", sucursal)

	if err := xreq.MakeRequest(
		ctx,
		baseurl,
		"/v1/cultivos",
		xreq.WithAuthorization(token),
		xreq.WithQueryParams(queryParams),
		xreq.WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return nil, err
	}

	return apiresponse.Data, nil
}
