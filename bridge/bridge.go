package bridge

import "github.com/sfperusacdev/identitysdk/bridge/variables"

type BridgeService struct {
	Variables *variables.VariablesService
}

func NewBridgeService(
	Variables *variables.VariablesService,
) *BridgeService {
	return &BridgeService{
		Variables: Variables,
	}
}
