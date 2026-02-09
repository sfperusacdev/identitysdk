package monitoring

import (
	"github.com/sfperusacdev/identitysdk/httpapi"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"monitoring",
	fx.Provide(NewMetricsService),
	fx.Provide(
		httpapi.AsRoute(NewHandlerNameHandler),
	),
)
