package telemetry

import (
	"net/http"
	"sync/atomic"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

type meterProviderSpy struct {
	metric.MeterProvider
	calls atomic.Int64
}

func (s *meterProviderSpy) Meter(name string, options ...metric.MeterOption) metric.Meter {
	s.calls.Add(1)
	return s.MeterProvider.Meter(name, options...)
}

func TestInstrumentationDoesNotUseGlobalMetrics(t *testing.T) {
	previous := otel.GetMeterProvider()
	spy := &meterProviderSpy{MeterProvider: metricnoop.NewMeterProvider()}
	otel.SetMeterProvider(spy)
	t.Cleanup(func() { otel.SetMeterProvider(previous) })
	service := &TelemetryService{tracerProvider: tracenoop.NewTracerProvider(), propagator: propagation.TraceContext{}}
	previousService := currentService.Swap(service)
	t.Cleanup(func() { currentService.Store(previousService) })
	service.httpClient(&http.Client{})
	service.HTTPMiddleware()
	GRPCClientHandler()
	GRPCServerHandler()
	if spy.calls.Load() != 0 {
		t.Fatal("transport instrumentation must not register global metrics")
	}
}
