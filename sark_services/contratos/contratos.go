package contratos

import (
	"context"
	"time"

	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
	"github.com/sfperusacdev/identitysdk/xreq"
)

type ContratosService struct {
	env bridgeidentity.IdentityProvider
}

func NewContratosService(
	env bridgeidentity.IdentityProvider,
) *ContratosService {
	return &ContratosService{
		env: env,
	}
}

type TrabajadorActivoDTO struct {
	TrabajadorCodigo string `json:"trabajador_codigo"`
	CodigoUnico      string `json:"codigo_unico"`
	TrabajadorName   string `json:"trabajador_name"`
}

// TrabajadoresActivos obtiene la lista de trabajadores activos. Si incluirInactivos
// es true, incluye también a los trabajadores inactivos.
func (s *ContratosService) TrabajadoresActivos(ctx context.Context, incluirInactivos bool) ([]TrabajadorActivoDTO, error) {
	apiURL, err := s.env.GetContratosServiceURL(ctx, s.env.Empresa(ctx))
	if err != nil {
		return nil, err
	}

	incluirInactivosParam := "no"
	if incluirInactivos {
		incluirInactivosParam = "yes"
	}

	var apiResponse struct {
		Data []TrabajadorActivoDTO `json:"data"`
	}
	if err := xreq.MakeRequest(ctx, apiURL,
		"/v1/trabajadores/lista-trabajadores-activos",
		xreq.WithQueryParam("incluir_inactivos", incluirInactivosParam),
		xreq.WithUnmarshalResponseInto(&apiResponse),
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
	); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

type ResumenTrabajadorFotocheckDTO struct {
	Codigo                    string     `json:"codigo"`
	DNI                       string     `json:"dni"`
	Nombres                   string     `json:"nombres"`
	ApellidoMaterno           string     `json:"apellido_materno"`
	ApellidoPaterno           string     `json:"apellido_paterno"`
	TipoDocumentoIdentidad    string     `json:"tipo_documento_identidad"`
	DocumentoIdentidad        string     `json:"documento_identidad"`
	CargoCodigo               string     `json:"cargo_codigo"`
	Cargo                     string     `json:"cargo"`
	PlanillaCodigo            string     `json:"planilla_codigo"`
	FotocheckWasPrintedAt     *time.Time `json:"fotocheck_was_printed_at"`
	ImageLocation             string     `json:"image_location"`
	Email                     string     `json:"email"`
	Telefono                  string     `json:"telefono"`
	Edad                      string     `json:"edad"`
	Sexo                      string     `json:"sexo"`
	Estado                    bool       `json:"estado"`
	TieneHuellas              bool       `json:"tiene_huellas"`
	FechaInicioUltimoContrato *time.Time `json:"fecha_inicio_ultimo_contrato"`
	FechaFinUltimoContrato    *time.Time `json:"fecha_fin_ultimo_contrato"`
	FechaIngreso              *time.Time `json:"fecha_ingreso"`
}

// ResumenTrabajadoresFotocheck obtiene los trabajadores para el listado de
// fotocheck. Si incluirInactivos es true, incluye también a los inactivos.
func (s *ContratosService) ResumenTrabajadoresFotocheck(ctx context.Context, incluirInactivos bool) ([]ResumenTrabajadorFotocheckDTO, error) {
	apiURL, err := s.env.GetContratosServiceURL(ctx, s.env.Empresa(ctx))
	if err != nil {
		return nil, err
	}

	incluirInactivosParam := "no"
	if incluirInactivos {
		incluirInactivosParam = "yes"
	}

	var apiResponse struct {
		Data []ResumenTrabajadorFotocheckDTO `json:"data"`
	}
	if err := xreq.MakeRequest(ctx, apiURL,
		"/v1/fotocheck/trabajadores/json",
		xreq.WithQueryParam("incluir_inactivos", incluirInactivosParam),
		xreq.WithUnmarshalResponseInto(&apiResponse),
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
	); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}
