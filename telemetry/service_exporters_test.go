package telemetry_test

import (
	"context"
	"log/slog"
	"net"
	"sync/atomic"
	"testing"

	"github.com/sfperusacdev/identitysdk/telemetry"
	collectorlog "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectormetric "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

type traceReceiver struct {
	collectortrace.UnimplementedTraceServiceServer
	called atomic.Bool
}

func (s *traceReceiver) Export(context.Context, *collectortrace.ExportTraceServiceRequest) (*collectortrace.ExportTraceServiceResponse, error) {
	s.called.Store(true)
	return &collectortrace.ExportTraceServiceResponse{}, nil
}

type metricReceiver struct {
	collectormetric.UnimplementedMetricsServiceServer
	called atomic.Bool
}

func (s *metricReceiver) Export(context.Context, *collectormetric.ExportMetricsServiceRequest) (*collectormetric.ExportMetricsServiceResponse, error) {
	s.called.Store(true)
	return &collectormetric.ExportMetricsServiceResponse{}, nil
}

type logReceiver struct {
	collectorlog.UnimplementedLogsServiceServer
	called atomic.Bool
}

func (s *logReceiver) Export(context.Context, *collectorlog.ExportLogsServiceRequest) (*collectorlog.ExportLogsServiceResponse, error) {
	s.called.Store(true)
	return &collectorlog.ExportLogsServiceResponse{}, nil
}

func TestGRPCExporters(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	server := grpc.NewServer()
	t.Cleanup(server.Stop)
	traces, metrics, logs := &traceReceiver{}, &metricReceiver{}, &logReceiver{}
	collectortrace.RegisterTraceServiceServer(server, traces)
	collectormetric.RegisterMetricsServiceServer(server, metrics)
	collectorlog.RegisterLogsServiceServer(server, logs)
	go func() { _ = server.Serve(listener) }()
	c := telemetry.DefaultConfig()
	c.Enabled, c.OTLP.Protocol, c.OTLP.Endpoint, c.Traces.SampleRatio = true, "grpc", "http://"+listener.Addr().String(), 1
	r, err := telemetry.NewTelemetryService(t.Context(), c, telemetry.ServiceInfo{Name: "grpc-export-test"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdown(t, r) })
	ctx, span := r.TracerProvider().Tracer("test").Start(t.Context(), "export over grpc")
	slog.InfoContext(ctx, "grpc export works")
	span.End()
	shutdown(t, r)
	if !traces.called.Load() || !logs.called.Load() {
		t.Fatalf("missing grpc signals: traces=%v logs=%v", traces.called.Load(), logs.called.Load())
	}
	if metrics.called.Load() {
		t.Fatal("metrics must not be exported")
	}
}
