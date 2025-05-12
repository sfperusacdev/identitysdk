package propertiesfx

import (
	"github.com/sfperusacdev/identitysdk/helpers/properties"
	"github.com/sfperusacdev/identitysdk/helpers/properties/properties_fx/handlers"
	propsprovider "github.com/sfperusacdev/identitysdk/helpers/properties/props_provider"
	"github.com/sfperusacdev/identitysdk/httpapi"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"system-properties",
	fx.Provide(
		fx.Private,
		fx.Annotate(
			propsprovider.NewSystemPropsPgProvider,
			fx.As(new(properties.SystemPropertiesMutator)),
		),
	),
	fx.Provide(
		httpapi.AsRoute(handlers.NewUpdatePropertiesHandler),
		httpapi.AsRoute(handlers.NewListarPropertiesHandler),
	),
)
