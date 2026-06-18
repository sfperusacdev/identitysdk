package bridge

import (
	"github.com/sfperusacdev/identitysdk/bridge/storage"
	"github.com/sfperusacdev/identitysdk/bridge/variables"
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
