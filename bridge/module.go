package bridge

import (
	bridgeidentity "github.com/sfperusacdev/identitysdk/bridge/identity"
	"github.com/sfperusacdev/identitysdk/bridge/variables"
	"go.uber.org/fx"
)

var Module = fx.Module("identitysdk/bridge",
	fx.Provide(
		fx.Annotate(
			bridgeidentity.NewDefaultProvider,
			fx.As(new(bridgeidentity.Provider)),
		),
		variables.NewGlobalVariablesService,
		variables.NewMeVariablesService,
		variables.NewVariablesService,
		NewBridgeService,
	),
)
