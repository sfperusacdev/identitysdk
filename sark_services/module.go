package sark_services

import (
	"github.com/sfperusacdev/identitysdk/sark_services/asistencia"
	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
	"github.com/sfperusacdev/identitysdk/sark_services/storage"
	"github.com/sfperusacdev/identitysdk/sark_services/variables"
	"go.uber.org/fx"
)

var Module = fx.Module("identitysdk/sark_services",
	fx.Provide(
		fx.Annotate(
			bridgeidentity.NewDefaultIdentityProvider,
			fx.As(new(bridgeidentity.IdentityProvider)),
		),
		bridgeidentity.NewIdentityService,
		variables.NewGlobalVariablesService,
		variables.NewMeVariablesService,
		variables.NewVariablesService,
		storage.NewStorageService,
		asistencia.NewAsistenciaService,
		NewSarkBridgeService,
	),
)
