package telemetry_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/sfperusacdev/identitysdk/telemetry"
	collectorlog "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	logpb "go.opentelemetry.io/proto/otlp/logs/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

type collector struct {
	mu        sync.Mutex
	spans     []*tracepb.Span
	logs      []*logpb.LogRecord
	resources []string
}

func newCollector(t *testing.T) (*collector, *httptest.Server) {
	t.Helper()
	c := &collector{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			w.WriteHeader(500)
			return
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		switch r.URL.Path {
		case "/v1/traces":
			var req collectortrace.ExportTraceServiceRequest
			if err := proto.Unmarshal(data, &req); err != nil {
				t.Error(err)
				w.WriteHeader(400)
				return
			}
			for _, resource := range req.ResourceSpans {
				c.resources = append(c.resources, resource.Resource.String())
				for _, scope := range resource.ScopeSpans {
					c.spans = append(c.spans, scope.Spans...)
				}
			}
		case "/v1/logs":
			var req collectorlog.ExportLogsServiceRequest
			if err := proto.Unmarshal(data, &req); err != nil {
				t.Error(err)
				w.WriteHeader(400)
				return
			}
			for _, resource := range req.ResourceLogs {
				for _, scope := range resource.ScopeLogs {
					c.logs = append(c.logs, scope.LogRecords...)
				}
			}
		default:
			t.Errorf("unexpected OTLP path %s", r.URL.Path)
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	t.Cleanup(srv.Close)
	return c, srv
}

func startTelemetryService(t *testing.T, endpoint string) *telemetry.TelemetryService {
	t.Helper()
	c := telemetry.DefaultConfig()
	c.Enabled, c.Environment, c.OTLP.Endpoint, c.Traces.SampleRatio = true, "test", endpoint, 1
	r, err := telemetry.NewTelemetryService(t.Context(), c, telemetry.ServiceInfo{Name: "test-service", Version: "1.2.3"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdown(t, r) })
	return r
}

func shutdown(t *testing.T, service *telemetry.TelemetryService) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := service.Shutdown(ctx); err != nil {
		t.Error(err)
	}
}
