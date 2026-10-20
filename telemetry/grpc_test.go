package telemetry_test

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/sfperusacdev/identitysdk/telemetry"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestTelemetryServiceGRPCPropagationAndAuthentication(t *testing.T) {
	col, srv := newCollector(t)
	r := startTelemetryService(t, srv.URL)
	listener := bufconn.Listen(1024 * 1024)
	t.Cleanup(func() { _ = listener.Close() })
	server := grpc.NewServer(grpc.StatsHandler(telemetry.GRPCServerHandler()), grpc.UnaryInterceptor(
		func(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
			md, _ := metadata.FromIncomingContext(ctx)
			if len(md.Get("x-access-token")) != 1 || md.Get("x-access-token")[0] != "RPC-SECRET" {
				return nil, status.Error(codes.Unauthenticated, "denied")
			}
			return next(ctx, req)
		}))
	healthpb.RegisterHealthServer(server, health.NewServer())
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithStatsHandler(telemetry.GRPCClientHandler()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	ctx, parent := r.TracerProvider().Tracer("test").Start(t.Context(), "rpc upstream")
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	client := healthpb.NewHealthClient(conn)
	if _, err := client.Check(ctx, &healthpb.HealthCheckRequest{}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("authentication lost: %v", err)
	}
	if _, err := client.Check(metadata.AppendToOutgoingContext(ctx, "x-access-token", "RPC-SECRET"), &healthpb.HealthCheckRequest{}); err != nil {
		t.Fatal(err)
	}
	watchCtx, cancelWatch := context.WithCancel(metadata.AppendToOutgoingContext(ctx, "x-access-token", "RPC-SECRET"))
	watch, err := client.Watch(watchCtx, &healthpb.HealthCheckRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := watch.Recv(); err != nil {
		t.Fatal(err)
	}
	cancelWatch()
	if _, err := watch.Recv(); status.Code(err) != codes.Canceled {
		t.Fatalf("unexpected stream cancellation: %v", err)
	}
	parent.End()
	server.GracefulStop()
	shutdown(t, r)
	col.mu.Lock()
	defer col.mu.Unlock()
	var servers, clients int
	for _, span := range col.spans {
		want := parent.SpanContext().TraceID()
		if !bytes.Equal(span.TraceId, want[:]) {
			t.Errorf("trace discontinuity: %s", span.Name)
		}
		if span.Kind == tracepb.Span_SPAN_KIND_SERVER {
			servers++
		}
		if span.Kind == tracepb.Span_SPAN_KIND_CLIENT {
			clients++
		}
		if strings.Contains(span.String(), "RPC-SECRET") {
			t.Error("gRPC credential exported")
		}
	}
	if servers != 3 || clients != 3 {
		t.Fatalf("RPC coverage servers=%d clients=%d", servers, clients)
	}
}
