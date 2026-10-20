package telemetry

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/stats"
)

func GRPCServerHandler() stats.Handler {
	service := Current()
	return otelgrpc.NewServerHandler(
		otelgrpc.WithTracerProvider(service.tracerProvider),
		otelgrpc.WithMeterProvider(noopMeterProvider),
		otelgrpc.WithPropagators(service.propagator),
	)
}

func GRPCClientHandler() stats.Handler {
	service := Current()
	return otelgrpc.NewClientHandler(
		otelgrpc.WithTracerProvider(service.tracerProvider),
		otelgrpc.WithMeterProvider(noopMeterProvider),
		otelgrpc.WithPropagators(service.propagator),
	)
}
