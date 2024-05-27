package services

import (
	"context"
	"log/slog"
	"net/url"
	"path"
	"strings"

	"github.com/sfperusacdev/identitysdk"
)

type ListDocumentoEstadoDto []struct {
	Estado string `json:"estado"`
	Peso   int64  `json:"peso"`
}

func (list ListDocumentoEstadoDto) Contains(s string) bool {
	for _, itm := range list {
		if itm.Estado == s {
			return true
		}
	}
	return false
}

func (s *ExternalBridgeService) GetEstadosDocumentoSegunTrabajador(
	ctx context.Context, documento, codigoTrabajador string) (ListDocumentoEstadoDto, error) {
	puesto, err := s.GetTrabajadorPosicion(ctx, codigoTrabajador)
	if err != nil {
		return nil, err
	}
	return s.GetEstadosDocumentoSegunPosicion(ctx, documento, puesto)
}

func (s *ExternalBridgeService) GetEstadosDocumentoSegunPosicion(
	ctx context.Context, documento, posicion string) (ListDocumentoEstadoDto, error) {
	posiciones, err := s.GetPuestosInferiores(ctx, posicion)
	if err != nil {
		return nil, err
	}

	var company, token = s.readCompanyAndToken(ctx)
	baseurl, err := identitysdk.GetGeneralServiceURL(ctx, company)
	if err != nil {
		slog.Error("error trying to retrieve `contratos` service url", "error", err)
		return nil, err
	}
	var apiresponse struct {
		Message string                 `json:"message"`
		Data    ListDocumentoEstadoDto `json:"data"`
	}
	var enpointPath = path.Join("/api/v1/documentos/", documento, "/estados")
	var values = url.Values{"puesto": []string{strings.Join(posiciones, ",")}}
	if err := s.makeRequestWithQueryPrams(ctx, baseurl, enpointPath, token, values, &apiresponse); err != nil {
		return nil, err
	}
	return apiresponse.Data, nil
}
