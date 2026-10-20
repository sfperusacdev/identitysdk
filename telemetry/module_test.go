package telemetry_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/sfperusacdev/identitysdk/telemetry"
	"go.uber.org/fx"
)

func TestTelemetryServiceFxShutdownOrder(t *testing.T) {
	col, srv := newCollector(t)
	r := startTelemetryService(t, srv.URL)
	app := fx.New(fx.NopLogger, fx.Supply(r), telemetry.Module,
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{OnStop: func(ctx context.Context) error {
				slog.InfoContext(ctx, "server drained before telemetry shutdown")
				return nil
			}})
		}))
	if err := app.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := app.Stop(t.Context()); err != nil {
		t.Fatal(err)
	}
	col.mu.Lock()
	defer col.mu.Unlock()
	for _, log := range col.logs {
		if log.Body.GetStringValue() == "server drained before telemetry shutdown" {
			return
		}
	}
	t.Fatal("server shutdown log was lost before exporters flushed")
}
