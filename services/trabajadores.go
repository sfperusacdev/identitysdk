package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/entities"
)

func (s *ExternalBridgeService) GetTrabajadores(ctx context.Context) ([]entities.ResumenTrabajadorDto, error) {
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
	if err := s.makeRequest(ctx, baseurl, enpointPath, token, &apiresponse); err != nil {
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
