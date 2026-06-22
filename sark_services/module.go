package sark_services

import (
	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
	"github.com/sfperusacdev/identitysdk/sark_services/storage"
	"github.com/sfperusacdev/identitysdk/sark_services/variables"
	"go.uber.org/fx"
)

var Module = fx.Module("identitysdk/sark_services",
	fx.Provide(
		fx.Annotate(
			bridgeidentity.NewDefaultProvider,
			fx.As(new(bridgeidentity.Provider)),
		),
		variables.NewGlobalVariablesService,
		variables.NewMeVariablesService,
		variables.NewVariablesService,
		storage.NewStorageService,
		NewBridgeService,
	),
)
