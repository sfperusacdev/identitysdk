package identity

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	integrationCache "github.com/sfperusacdev/identitysdk/internal/integracion_cache"
	"github.com/sfperusacdev/identitysdk/xreq"
	"github.com/user0608/ifdevmode"
)

// IntegracionExternaCodigo devuelve el codigo de la compania en el sistema externo.
func (s *IdentityService) IntegracionExternaCodigo(ctx context.Context, companyCode string) (string, error) {
	if debugValue := os.Getenv("DEBUG_OVERRIDE_INTEGRATION_EXTERNAL_CODE"); debugValue != "" {
		return debugValue, nil
	}

	if cachedState := integrationCache.DefaultCache.Get(ctx, companyCode); cachedState != nil {
		if ifdevmode.Yes() {
			slog.Info("IntegracionExternaCodigo read from cache")
		}
		return cachedState.ExternalReff, nil
	}

	var apiResponse struct {
		Data integrationCache.IntegracionState `json:"data"`
	}
	if err := xreq.MakeRequest(ctx,
		s.identity.IdentityServer(),
		fmt.Sprintf("/v1/get-external-info-empresa/%s", companyCode),
		xreq.WithUnmarshalResponseInto(&apiResponse),
	); err != nil {
		return "", err
	}

	integrationCache.DefaultCache.Set(ctx, companyCode, apiResponse.Data)
	return strings.TrimSpace(apiResponse.Data.ExternalReff), nil
}

// IntegracionExternaSucursalCodigo devuelve el codigo de la sucursal en el sistema externo.
func (s *IdentityService) IntegracionExternaSucursalCodigo(ctx context.Context, companyCode, branchCode string) (string, error) {
	if debugValue := os.Getenv("DEBUG_OVERRIDE_INTEGRATION_EXTERNAL_BRANCH_CODE"); debugValue != "" {
		return debugValue, nil
	}

	var apiResponse struct {
		Data struct {
			ExternalReff string `json:"external_reff"`
		} `json:"data"`
	}
	if err := xreq.MakeRequest(ctx,
		s.identity.IdentityServer(),
		fmt.Sprintf("/v1/get-external-info-empresa-sucursal/%s/%s", companyCode, branchCode),
		xreq.WithUnmarshalResponseInto(&apiResponse),
	); err != nil {
		return "", err
	}
	return strings.TrimSpace(apiResponse.Data.ExternalReff), nil
}

func splitIntegrationURL(value string) (string, bool) {
	const suffix = ":ro"
	value = strings.TrimSpace(value)
	if before, ok := strings.CutSuffix(value, suffix); ok {
		return strings.TrimRight(before, "/"), true
	}
	return strings.TrimRight(value, "/"), false
}

// IntegracionExternaURL devuelve la URL del servicio de la compania en el sistema externo.
// El segundo valor indica si el servicio de integracion es de solo lectura.
func (s *IdentityService) IntegracionExternaURL(ctx context.Context, companyCode string) (integrationURL string, readOnly bool, err error) {
	if debugValue := os.Getenv("DEBUG_OVERRIDE_INTEGRATION_EXTERNAL_URL"); debugValue != "" {
		integrationURL, readOnly := splitIntegrationURL(debugValue)
		return integrationURL, readOnly, nil
	}

	if cachedState := integrationCache.DefaultCache.Get(ctx, companyCode); cachedState != nil {
		if ifdevmode.Yes() {
			slog.Info("IntegracionExternaURL read from cache")
		}
		integrationURL, readOnly := splitIntegrationURL(cachedState.IntegrationURL)
		return integrationURL, readOnly, nil
	}

	var apiResponse struct {
		Data integrationCache.IntegracionState `json:"data"`
	}
	if err := xreq.MakeRequest(ctx,
		s.identity.IdentityServer(),
		fmt.Sprintf("/v1/get-external-info-empresa/%s", companyCode),
		xreq.WithUnmarshalResponseInto(&apiResponse),
	); err != nil {
		return "", false, err
	}

	integrationCache.DefaultCache.Set(ctx, companyCode, apiResponse.Data)
	integrationURL, readOnly = splitIntegrationURL(apiResponse.Data.IntegrationURL)
	return integrationURL, readOnly, nil
}

func (s *IdentityService) IntegracionExternaCodigo_(ctx context.Context) (string, error) {
	return s.IntegracionExternaCodigo(ctx, s.identity.Empresa(ctx))
}

func (s *IdentityService) IntegracionExternaSucursalCodigo_(ctx context.Context) (string, error) {
	companyCode, branchCode := s.identity.Empresa_Sucursal(ctx)
	return s.IntegracionExternaSucursalCodigo(ctx, companyCode, branchCode)
}

func (s *IdentityService) IntegracionExternaURL_(ctx context.Context) (integrationURL string, readOnly bool, err error) {
	return s.IntegracionExternaURL(ctx, s.identity.Empresa(ctx))
}
