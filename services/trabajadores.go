package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/entities"
	"github.com/sfperusacdev/identitysdk/xreq"
	"github.com/user0608/goones/errs"
)

func (s *ExternalBridgeService) GetTrabajadores(ctx context.Context, incluirInactivos bool) ([]entities.ResumenTrabajadorDto, error) {
	var company, token = s.readCompanyAndToken(ctx)
	baseurl, err := identitysdk.GetContratosServiceURL(ctx, company)
	if err != nil {
		slog.Error("error trying to retrieve `contratos` service url", "error", err)
		return nil, err
	}
	var apiresponse struct {
		Message string `json:"message"`
		Data    []struct {
			Codigo                 string     `json:"codigo"`
			Dni                    string     `json:"dni"`
			Nombres                string     `json:"nombres"`
			ApellidoMaterno        string     `json:"apellido_materno"`
			ApellidoPaterno        string     `json:"apellido_paterno"`
			TipoDocumentoIdentidad string     `json:"tipo_documento_identidad"`
			DocumentoIdentidad     string     `json:"documento_identidad"`
			CargoCodigo            string     `json:"cargo_codigo"`
			Cargo                  string     `json:"cargo"`
			PlanillaCodigo         string     `json:"planilla_codigo"`
			FechaIngreso           string     `json:"fecha_ingreso"`
			ImageLocation          string     `json:"image_location"`
			Email                  string     `json:"email"`
			Telefono               string     `json:"telefono"`
			Edad                   string     `json:"edad"`
			Estado                 bool       `json:"estado"`
			Sexo                   string     `json:"sexo"`
			DeletedAt              *time.Time `json:"deleted_at"`
			IsDisabled             bool       `json:"is_disabled"`
		} `json:"data"`
	}
	var enpointPath = "/v1/fotocheck/trabajadores/json"

	var queryParams = make(url.Values)
	if incluirInactivos {
		queryParams.Set("incluir_inactivos", "yes")
	}

	if err := xreq.MakeRequest(ctx,
		baseurl, enpointPath,
		xreq.WithAuthorization(token),
		xreq.WithUnmarshalResponseInto(&apiresponse),
		xreq.WithQueryParams(queryParams),
	); err != nil {
		return nil, err
	}

	var response = make([]entities.ResumenTrabajadorDto, 0, len(apiresponse.Data))
	for _, itm := range apiresponse.Data {
		response = append(response, entities.ResumenTrabajadorDto{
			Codigo:          itm.Codigo,
			Nombres:         itm.Nombres,
			ApellidoPaterno: itm.ApellidoPaterno,
			ApellidoMaterno: itm.ApellidoMaterno,
			Documento: entities.ResumenTrabajadorDocumentoIdentidadDto{
				Tipo:   itm.TipoDocumentoIdentidad,
				Numero: itm.DocumentoIdentidad,
			},
			Cargo: entities.ResumenTrabajadorCargoDto{
				Codigo:      itm.CargoCodigo,
				Descripcion: itm.Cargo,
			},
			PlanillaCodigo: itm.PlanillaCodigo,
			FechaIngreso:   itm.FechaIngreso,
			Email:          itm.Email,
			Telefono:       itm.Telefono,
			Sexo:           itm.Sexo,
			DeletedAt:      itm.DeletedAt,
		})
	}
	return response, nil
}

type FaztCreateTrabajadorDTO struct {
	Codigo          string  `json:"codigo"`
	Nombres         string  `json:"nombres"`
	ApellidoMaterno string  `json:"apellido_materno"`
	ApellidoPaterno string  `json:"apellido_paterno"`
	PlanillaCodigo  *string `json:"planilla_codigo"`
	Email           *string `json:"email"`
	Telefono        *string `json:"telefono"`
	Sexo            *string `json:"sexo"`
}

func (s *ExternalBridgeService) FastImportTrabajadores(ctx context.Context, trabajadores []FaztCreateTrabajadorDTO) error {
	if len(trabajadores) == 0 {
		return nil
	}
	company, token := s.readCompanyAndToken(ctx)
	var requestBuff bytes.Buffer
	bodyEncoder := json.NewEncoder(&requestBuff)
	if err := bodyEncoder.Encode(trabajadores); err != nil {
		slog.Error("error encoding trabajadores to JSON", "error", err, "company", company)
		return errs.BadRequestDirect("no se pudo procesar la solicitud para importar trabajadores")
	}

	baseurl, err := identitysdk.GetContratosServiceURL(ctx, company)
	if err != nil {
		slog.Error("error retrieving contratos service URL", "error", err, "company", company)
		return errs.BadRequestDirect("no se pudo obtener la URL del servicio de contratos")
	}
	if err := xreq.MakeRequest(ctx,
		baseurl,
		"/v1/trabajadores/fazt/import",
		xreq.WithMethod(http.MethodPost),
		xreq.WithAuthorization(token),
		xreq.WithRequestBody(&requestBuff),
		xreq.WithJsonContentType(),
	); err != nil {
		slog.Error("error making request to contratos service", "error", err, "company", company, "endpoint", "/v1/trabajadores/fazt/import")
		return errs.BadRequestDirect("no se pudo importar trabajadores")
	}
	return nil
}

// public user session requiered
func (s *ExternalBridgeService) GetMyInfo(ctx context.Context, empresa string) (*entities.TrabajadorDto, error) {
	var token = identitysdk.Token(ctx)
	baseurl, err := identitysdk.GetContratosServiceURL(ctx, empresa)
	if err != nil {
		slog.Error("error trying to retrieve `contratos` service url", "error", err)
		return nil, err
	}
	// var mapa = map[string]any{}
	var apiresponse struct {
		Message string                   `json:"message"`
		Data    []entities.TrabajadorDto `json:"data"`
	}
	var enpointPath = fmt.Sprintf("/v1/public/resumen/trabajador/info/%s", empresa)
	if err := xreq.MakeRequest(ctx,
		baseurl, enpointPath,
		xreq.WithAuthorization(token),
		xreq.WithJsonContentType(),
		xreq.WithUnmarshalResponseInto(&apiresponse),
	); err != nil {
		return nil, err
	}
	if len(apiresponse.Data) == 0 {
		return nil, errs.BadRequestDirect("No se encontraron datos para el usuario o trabajador especificado.")
	}
	return &apiresponse.Data[0], nil
}
