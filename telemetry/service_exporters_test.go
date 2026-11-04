package telemetry_test

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/sfperusacdev/identitysdk/telemetry"
	collectorlog "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectormetric "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type traceReceiver struct {
	collectortrace.UnimplementedTraceServiceServer
	called   atomic.Bool
	withAuth atomic.Bool
}

func (s *traceReceiver) Export(ctx context.Context, _ *collectortrace.ExportTraceServiceRequest) (*collectortrace.ExportTraceServiceResponse, error) {
	s.called.Store(true)
	if md, ok := metadata.FromIncomingContext(ctx); ok && len(md.Get("authorization")) == 1 && md.Get("authorization")[0] == "Basic placeholder-grpc" && len(md.Get("organization")) == 1 && md.Get("organization")[0] == "default" {
		s.withAuth.Store(true)
	}
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
	called   atomic.Bool
	withAuth atomic.Bool
}

func (s *logReceiver) Export(ctx context.Context, _ *collectorlog.ExportLogsServiceRequest) (*collectorlog.ExportLogsServiceResponse, error) {
	s.called.Store(true)
	if md, ok := metadata.FromIncomingContext(ctx); ok && len(md.Get("authorization")) == 1 && md.Get("authorization")[0] == "Basic placeholder-grpc" && len(md.Get("organization")) == 1 && md.Get("organization")[0] == "default" {
		s.withAuth.Store(true)
	}
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
	c.OTLP.Headers = map[string]string{"authorization": "Basic placeholder-grpc", "organization": "default"}
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
	if !traces.withAuth.Load() || !logs.withAuth.Load() {
		t.Fatal("gRPC exporters did not send configured headers")
	}
	if metrics.called.Load() {
		t.Fatal("metrics must not be exported")
	}
}

func TestHTTPExporterHeaders(t *testing.T) {
	t.Setenv("OPENOBSERVE_AUTH_HEADER", "Basic placeholder-config")
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "Authorization=Basic%20placeholder-env,stream-name=env-stream")
	for _, tc := range []struct {
		name       string
		headers    map[string]string
		wantAuth   string
		wantStream string
	}{
		{"configured headers override environment", map[string]string{
			"Authorization": "${OPENOBSERVE_AUTH_HEADER}", "stream-name": "configured-stream",
		}, "Basic placeholder-config", "configured-stream"},
		{"environment fallback", nil, "Basic placeholder-env", "env-stream"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var traces, logs atomic.Bool
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != tc.wantAuth || r.Header.Get("stream-name") != tc.wantStream {
					t.Errorf("unexpected exporter headers on %s", r.URL.Path)
				}
				switch r.URL.Path {
				case "/api/default/v1/traces":
					traces.Store(true)
				case "/api/default/v1/logs":
					logs.Store(true)
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/x-protobuf")
			}))
			defer srv.Close()
			c := telemetry.DefaultConfig()
			c.Enabled, c.OTLP.Endpoint, c.Traces.SampleRatio = true, srv.URL+"/api/default", 1
			c.OTLP.Headers = tc.headers
			r, err := telemetry.NewTelemetryService(t.Context(), c, telemetry.ServiceInfo{Name: "header-test"})
			if err != nil {
				t.Fatal(err)
			}
			ctx, span := r.TracerProvider().Tracer("test").Start(t.Context(), "export headers")
			slog.InfoContext(ctx, "export headers")
			span.End()
			shutdown(t, r)
			if !traces.Load() || !logs.Load() {
				t.Fatalf("missing HTTP signals: traces=%v logs=%v", traces.Load(), logs.Load())
			}
		})
	}
}
