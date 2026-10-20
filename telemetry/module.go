package telemetry

import (
	"go.uber.org/fx"
)

// Register before server hooks so exporters stop after requests drain.
var Module = fx.Module("telemetry", fx.Invoke(registerLifecycle))

func registerLifecycle(lc fx.Lifecycle, service *TelemetryService) {
	lc.Append(fx.Hook{OnStop: service.Shutdown})
}
