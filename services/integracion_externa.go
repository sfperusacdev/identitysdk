package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/sfperusacdev/identitysdk"
	integracioncache "github.com/sfperusacdev/identitysdk/internal/integracion_cache"
	"github.com/user0608/ifdevmode"
)

// IntegracionExternaCodigo devuelve el codigo de la compania en el sistema externo
func (s *ExternalBridgeService) IntegracionExternaCodigo(ctx context.Context, company string) (string, error) {
	{ // variable de entorno, solo para propositos de debug
		var debugValue = os.Getenv("DEBUG_OVERRIDE_INTEGRATION_EXTERNAL_CODE")
		if debugValue != "" {
			return debugValue, nil
		}
	}
	var cachedValue = integracioncache.DefaultCache.Get(ctx, company)
	if cachedValue != nil {
		if ifdevmode.Yes() {
			slog.Info("IntegracionExternaCodigo read from cache")
		}
		return cachedValue.ExternalReff, nil
	}
	var baseUrl = identitysdk.GetIdentityServer()
	var enpointPath = fmt.Sprintf("/v1/get-external-info-empresa/%s", company)
	var apiresponse struct {
		Message string                            `json:"message"`
		Data    integracioncache.IntegracionState `json:"data"`
	}
	if err := s.makeRequest(ctx, baseUrl, enpointPath, "-", &apiresponse); err != nil {
		return "", err
	}
	integracioncache.DefaultCache.Set(ctx, company, apiresponse.Data)
	return apiresponse.Data.ExternalReff, nil
}

func (s *ExternalBridgeService) integracionExternaURlSplit(val string) (string, bool) {
	var values = strings.Split(val, ":")
	if len(values) == 2 {
		if values[1] == "ro" {
			return strings.TrimRight(values[0], "/"), true
		}
	}
	return strings.TrimRight(val, "/"), false
}

// IntegracionExternaCodigo devuelve la url del servicio de la compania en el sistema externo
// devuleve true en el segundo campo para indicar que el servicio de integracion es de solo lecutra
func (s *ExternalBridgeService) IntegracionExternaURl(ctx context.Context, company string) (integrationURL string, readOnly bool, err error) {
	{ // variable de entorno, solo para propositos de debug
		var debugValue = os.Getenv("DEBUG_OVERRIDE_INTEGRATION_EXTERNAL_URL")
		if debugValue != "" {
			value, readOnly := s.integracionExternaURlSplit(debugValue)
			return value, readOnly, nil
		}
	}
	var cachedValue = integracioncache.DefaultCache.Get(ctx, company)
	if cachedValue != nil {
		if ifdevmode.Yes() {
			slog.Info("IntegracionExternaURl read from cache")
		}
		val, readOnly := s.integracionExternaURlSplit(cachedValue.IntegrationURL)
		return val, readOnly, nil
	}
	var baseUrl = identitysdk.GetIdentityServer()
	var enpointPath = fmt.Sprintf("/v1/get-external-info-empresa/%s", company)
	var apiresponse struct {
		Message string                            `json:"message"`
		Data    integracioncache.IntegracionState `json:"data"`
	}
	if err := s.makeRequest(ctx, baseUrl, enpointPath, "-", &apiresponse); err != nil {
		return "", false, err
	}
	integracioncache.DefaultCache.Set(ctx, company, apiresponse.Data)
	val, readOnly := s.integracionExternaURlSplit(apiresponse.Data.IntegrationURL)
	return val, readOnly, nil
}
