package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/xreq"
)

func (s *ExternalBridgeService) GetTrabajadorPosicion(ctx context.Context, codigo string) (string, error) {
	var company, token = s.readCompanyAndToken(ctx)
	baseurl, err := identitysdk.GetContratosServiceURL(ctx, company)
	if err != nil {
		slog.Error("error trying to retrieve `contratos` service url", "error", err)
		return "", err
	}
	var enpointPath = fmt.Sprintf("/v1/fotocheck/%s/json", codigo)
	var apiresponse struct {
		Message string `json:"message"`
		Data    []struct {
			CargoCodigo string `json:"cargo_codigo"`
		} `json:"data"`
	}
	if err := xreq.MakeRequest(ctx,
		baseurl, enpointPath,
		xreq.WithAuthorization(token),
		xreq.WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return "", err
	}
	if len(apiresponse.Data) == 0 {
		return "", ErrNotFound
	}
	return apiresponse.Data[0].CargoCodigo, nil
}
