package services

import (
	"context"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/configs"
	"github.com/sfperusacdev/identitysdk/sark_services/contratos"
)

// Deprecated: use contratos.ErrNotFound instead.
var ErrNotFound = contratos.ErrNotFound

type ExternalBridgeService struct {
	configProvider configs.GeneralServiceConfigProvider
}

func NewExternalBridgeService(
	configProvider configs.GeneralServiceConfigProvider,
) *ExternalBridgeService {
	return &ExternalBridgeService{
		configProvider: configProvider,
	}
}

func (*ExternalBridgeService) readCompanyAndToken(ctx context.Context) (string, string) {
	var company = identitysdk.Empresa(ctx)
	var token = identitysdk.Token(ctx)
	return company, token
}
