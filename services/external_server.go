package services

import (
	"context"
	"errors"

	"github.com/sfperusacdev/identitysdk"
)

var ErrNotFound = errors.New("the record you are looking for was not found")

type ExternalBridgeService struct{}

func NewExternalBridgeService() *ExternalBridgeService {
	return &ExternalBridgeService{}
}

func (*ExternalBridgeService) readCompanyAndToken(ctx context.Context) (string, string) {
	var company = identitysdk.Empresa(ctx)
	var token = identitysdk.Token(ctx)
	return company, token
}
