package sark_services

import (
	"github.com/sfperusacdev/identitysdk/sark_services/storage"
	"github.com/sfperusacdev/identitysdk/sark_services/variables"
)

type BridgeService struct {
	Variables *variables.VariablesService
	Storage   *storage.StorageService
}

func NewBridgeService(
	Variables *variables.VariablesService,
	Storage *storage.StorageService,
) *BridgeService {
	return &BridgeService{
		Variables: Variables,
		Storage:   Storage,
	}
}
