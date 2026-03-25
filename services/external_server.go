package services

import (
	"context"
	"errors"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/configs"
)

var ErrNotFound = errors.New("the record you are looking for was not found")

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
