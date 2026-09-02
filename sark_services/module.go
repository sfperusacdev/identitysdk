package sark_services

import (
	"github.com/sfperusacdev/identitysdk/sark_services/asistencia"
	"github.com/sfperusacdev/identitysdk/sark_services/global"
	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
	"github.com/sfperusacdev/identitysdk/sark_services/storage"
	"github.com/sfperusacdev/identitysdk/sark_services/variables"
	"go.uber.org/fx"
)

var Module = fx.Module("identitysdk/sark_services",
	fx.Provide(
		bridgeidentity.NewIdentityProvider,
		bridgeidentity.NewIdentityService,
		variables.NewGlobalVariablesService,
		variables.NewMeVariablesService,
		variables.NewVariablesService,
		storage.NewStorageService,
		asistencia.NewAsistenciaService,
		global.NewGlobalService,
		NewSarkBridgeService,
	),
)
