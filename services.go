package identitysdk

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type queryServiceURLResponse struct {
	Message string `json:"message"`
	Data    struct {
		Location string `json:"location"`
	} `json:"data"`
}

func GetServiceURL(ctx context.Context, companyCode string, resourceCode string) (string, error) {
	var queryParams = url.Values{}
	queryParams.Set("company_code", companyCode)
	queryParams.Set("resource_code", resourceCode)
	hostUrl, err := url.JoinPath(identityAddress, "/api/v1/get-service-location")

	if err != nil {
		slog.Error("QueryServiceURL: error JoinPath", "error", err,
			"company", companyCode, "resource", resourceCode)
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hostUrl, nil)
	if err != nil {
		slog.Error("QueryServiceURL: error NewRequestWithContext", "error", err,
			"company", companyCode, "resource", resourceCode)
		return "", err
	}
	req.URL.RawQuery = queryParams.Encode()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("QueryServiceURL: error Doing Request", "error", err,
			"company", companyCode, "resource", resourceCode)
		return "", err
	}
	defer res.Body.Close()
	var data queryServiceURLResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		slog.Error("QueryServiceURL: Error decoding json response", "error", err,
			"company", companyCode, "resource", resourceCode)
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		slog.Warn("QueryServiceURL: status code != 200", "status", res.Status,
			"company", companyCode, "resource", resourceCode, "identity-response", data.Message)
		return "", errors.New("expected status code 200 on service request")
	}
	var location = strings.TrimSpace(data.Data.Location)
	if location == "" {
		slog.Warn("QueryServiceURL: empty location", "status", res.Status,
			"company", companyCode, "resource", resourceCode)
		return "", errors.New("service not found")
	}
	return location, nil
}

const (
	almacen_service      string = "com.sfperusac.almacen"
	sfsire_service       string = "com.sfperusac.sfsire"
	asistencia_service   string = "com.sfperusac.asistencia"
	syncdata_service     string = "com.sfperusac.syncdata"
	contratos_service    string = "com.sfperusac.contratos"
	general_service      string = "com.sfperusac.general"
	tareoapp_service     string = "com.sfperusac.tareoapp"
	whatsapp_api_service string = "con.sfperusac.whatsapp_api"
	mensajeria_service   string = "com.sfperusac.mensajeria"
)

func GetAlmacenServiceURL(ctx context.Context, companyCode string) (string, error) {
	return GetServiceURL(ctx, companyCode, almacen_service)
}

func GetSireServiceURL(ctx context.Context, companyCode string) (string, error) {
	return GetServiceURL(ctx, companyCode, sfsire_service)
}

func GetAsistenciaServiceURL(ctx context.Context, companyCode string) (string, error) {
	return GetServiceURL(ctx, companyCode, asistencia_service)
}

func GetSyncdataServiceURL(ctx context.Context, companyCode string) (string, error) {
	return GetServiceURL(ctx, companyCode, syncdata_service)
}

func GetContratosServiceURL(ctx context.Context, companyCode string) (string, error) {
	return GetServiceURL(ctx, companyCode, contratos_service)
}

func GetTareoServiceURL(ctx context.Context, companyCode string) (string, error) {
	return GetServiceURL(ctx, companyCode, tareoapp_service)
}

func GetGeneralServiceURL(ctx context.Context, companyCode string) (string, error) {
	return GetServiceURL(ctx, companyCode, general_service)
}

func GetWhatsAppApiServiceURL(ctx context.Context, companyCode string) (string, error) {
	return GetServiceURL(ctx, companyCode, whatsapp_api_service)
}

func GetMensajeriaServiceURL(ctx context.Context, companyCode string) (string, error) {
	return GetServiceURL(ctx, companyCode, mensajeria_service)
}

func Tz(ctx context.Context) (*time.Location, error) {
	claims, ok := JwtClaims(ctx)
	if !ok {
		return nil, errors.New("sesión inválida, contexto no encontrado")
	}
	if claims.Zona == "" {
		return nil, errors.New("zona horaria no definida para este dominio")
	}
	location, err := time.LoadLocation(claims.Zona)
	if err != nil {
		slog.Error("load location", "error", err, "tz", claims.Zona)
		return nil, err
	}
	return location, nil
}
