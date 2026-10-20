package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

type ServiceInfo struct {
	Name    string
	Version string
}

type TelemetryService struct {
	config          Config
	tracerProvider  trace.TracerProvider
	propagator      propagation.TextMapPropagator
	shutdownHooks   []func(context.Context) error
	shutdownOnce    sync.Once
	shutdownErr     error
	instrumentedDBs sync.Map
	restoreGlobals  func()
}

var disabledService = &TelemetryService{
	tracerProvider: tracenoop.NewTracerProvider(),
	propagator:     propagation.NewCompositeTextMapPropagator(),
}

var currentService atomic.Pointer[TelemetryService]

func Current() *TelemetryService {
	if service := currentService.Load(); service != nil {
		return service
	}
	return disabledService
}

func (s *TelemetryService) Enabled() bool { return s.config.Enabled }

func (s *TelemetryService) TracerProvider() trace.TracerProvider { return s.tracerProvider }

func (s *TelemetryService) Propagator() propagation.TextMapPropagator { return s.propagator }

func NewTelemetryService(ctx context.Context, c Config, serviceInfo ServiceInfo) (_ *TelemetryService, err error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	s := &TelemetryService{
		config:         c,
		tracerProvider: tracenoop.NewTracerProvider(),
		propagator:     propagation.NewCompositeTextMapPropagator(),
	}
	defer func() {
		if err != nil {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.Shutdown(cleanupCtx)
		}
	}()
	var bridge slog.Handler
	if c.Enabled {
		bridge, err = s.initializeSignals(ctx, serviceInfo)
		if err != nil {
			return nil, err
		}
	}
	s.activate(bridge)
	return s, nil
}

func (s *TelemetryService) initializeSignals(ctx context.Context, serviceInfo ServiceInfo) (slog.Handler, error) {
	c := s.config
	res, err := newTelemetryResource(ctx, serviceInfo.Name, serviceInfo.Version, c.Environment)
	if err != nil {
		return nil, err
	}
	s.propagator = propagation.TraceContext{}
	if c.Traces.Enabled {
		provider, err := newTracerProvider(ctx, c, res)
		if err != nil {
			return nil, err
		}
		s.tracerProvider = provider
		s.shutdownHooks = append(s.shutdownHooks, provider.Shutdown)
	}
	if c.Logs.ExportOTLP {
		provider, bridge, err := newLoggerProvider(ctx, c, res)
		if err != nil {
			return nil, err
		}
		s.shutdownHooks = append(s.shutdownHooks, provider.Shutdown)
		return bridge, nil
	}
	return nil, nil
}

func (s *TelemetryService) activate(bridge slog.Handler) {
	previousLogger := slog.Default()
	previousTracer, previousPropagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	previousErrorHandler := otel.GetErrorHandler()
	previousService := currentService.Swap(s)
	otel.SetErrorHandler(newExportErrorHandler(os.Stdout))
	otel.SetTracerProvider(s.tracerProvider)
	otel.SetTextMapPropagator(s.propagator)
	loggerConfig := s.config
	if !s.Enabled() {
		loggerConfig = DefaultConfig()
	}
	slog.SetDefault(newTelemetryLogger(loggerConfig, os.Stdout, bridge))
	s.restoreGlobals = func() {
		if currentService.CompareAndSwap(s, previousService) {
			slog.SetDefault(previousLogger)
			otel.SetTracerProvider(previousTracer)
			otel.SetTextMapPropagator(previousPropagator)
			otel.SetErrorHandler(previousErrorHandler)
		}
	}
}

func (s *TelemetryService) Shutdown(ctx context.Context) error {
	s.shutdownOnce.Do(func() {
		for _, stop := range s.shutdownHooks {
			s.shutdownErr = errors.Join(s.shutdownErr, stop(ctx))
		}
		if s.restoreGlobals != nil {
			s.restoreGlobals()
		}
	})
	return s.shutdownErr
}
