package identitysdk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	OVR_ALMACEN      = "UNSAFE_DEV_ONLY_OVERRIDE_ALMACEN_URL"
	OVR_SIRE         = "UNSAFE_DEV_ONLY_OVERRIDE_SIRE_URL"
	OVR_ASISTENCIA   = "UNSAFE_DEV_ONLY_OVERRIDE_ASISTENCIA_URL"
	OVR_SYNCDATA     = "UNSAFE_DEV_ONLY_OVERRIDE_SYNCDATA_URL"
	OVR_CONTRATOS    = "UNSAFE_DEV_ONLY_OVERRIDE_CONTRATOS_URL"
	OVR_TAREO        = "UNSAFE_DEV_ONLY_OVERRIDE_TAREO_URL"
	OVR_GENERAL      = "UNSAFE_DEV_ONLY_OVERRIDE_GENERAL_URL"
	OVR_WHATSAPP_API = "UNSAFE_DEV_ONLY_OVERRIDE_WHATSAPP_API_URL"
	OVR_MENSAJERIA   = "UNSAFE_DEV_ONLY_OVERRIDE_MENSAJERIA_URL"
)

var overrides = map[string]string{}

func init() {
	load := func(key string) {
		v := strings.TrimSpace(os.Getenv(key))
		if v != "" {
			overrides[key] = v
		}
	}

	load(OVR_ALMACEN)
	load(OVR_SIRE)
	load(OVR_ASISTENCIA)
	load(OVR_SYNCDATA)
	load(OVR_CONTRATOS)
	load(OVR_TAREO)
	load(OVR_GENERAL)
	load(OVR_WHATSAPP_API)
	load(OVR_MENSAJERIA)
}

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
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hostUrl, nil)
	if err != nil {
		return "", err
	}

	req.URL.RawQuery = queryParams.Encode()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var data queryServiceURLResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", errors.New("expected status code 200 on service request")
	}

	location := strings.TrimSpace(data.Data.Location)
	if location == "" {
		return "", errors.New("service not found")
	}

	return location, nil
}

func overrideOrFetch(ctx context.Context, companyCode string, envVar string, resource string) (string, error) {
	if v, ok := overrides[envVar]; ok {
		_, err := url.ParseRequestURI(v)
		if err != nil {
			return "", errors.New("invalid override url: " + envVar)
		}
		return v, nil
	}
	return GetServiceURL(ctx, companyCode, resource)
}

const (
	almacen_service      = "com.sfperusac.almacen"
	sfsire_service       = "com.sfperusac.sfsire"
	asistencia_service   = "com.sfperusac.asistencia"
	syncdata_service     = "com.sfperusac.syncdata"
	contratos_service    = "com.sfperusac.contratos"
	general_service      = "com.sfperusac.general"
	tareoapp_service     = "com.sfperusac.tareoapp"
	whatsapp_api_service = "con.sfperusac.whatsapp_api"
	mensajeria_service   = "com.sfperusac.mensajeria"
)

func GetAlmacenServiceURL(ctx context.Context, companyCode string) (string, error) {
	return overrideOrFetch(ctx, companyCode, OVR_ALMACEN, almacen_service)
}

func GetSireServiceURL(ctx context.Context, companyCode string) (string, error) {
	return overrideOrFetch(ctx, companyCode, OVR_SIRE, sfsire_service)
}

func GetAsistenciaServiceURL(ctx context.Context, companyCode string) (string, error) {
	return overrideOrFetch(ctx, companyCode, OVR_ASISTENCIA, asistencia_service)
}

func GetSyncdataServiceURL(ctx context.Context, companyCode string) (string, error) {
	return overrideOrFetch(ctx, companyCode, OVR_SYNCDATA, syncdata_service)
}

func GetContratosServiceURL(ctx context.Context, companyCode string) (string, error) {
	return overrideOrFetch(ctx, companyCode, OVR_CONTRATOS, contratos_service)
}

func GetTareoServiceURL(ctx context.Context, companyCode string) (string, error) {
	return overrideOrFetch(ctx, companyCode, OVR_TAREO, tareoapp_service)
}

func GetGeneralServiceURL(ctx context.Context, companyCode string) (string, error) {
	return overrideOrFetch(ctx, companyCode, OVR_GENERAL, general_service)
}

func GetWhatsAppApiServiceURL(ctx context.Context, companyCode string) (string, error) {
	return overrideOrFetch(ctx, companyCode, OVR_WHATSAPP_API, whatsapp_api_service)
}

func GetMensajeriaServiceURL(ctx context.Context, companyCode string) (string, error) {
	return overrideOrFetch(ctx, companyCode, OVR_MENSAJERIA, mensajeria_service)
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
		return nil, err
	}
	return location, nil
}
