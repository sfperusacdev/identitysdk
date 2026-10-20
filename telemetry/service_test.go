package telemetry_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sfperusacdev/identitysdk/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestTelemetryServiceSignalSwitches(t *testing.T) {
	for _, tt := range []struct {
		name         string
		traces, logs bool
	}{
		{"traces without logs", true, false},
		{"logs without traces", false, true},
		{"stdout only", false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			col, srv := newCollector(t)
			c := telemetry.DefaultConfig()
			c.Enabled, c.OTLP.Endpoint, c.Traces.SampleRatio = true, srv.URL, 1
			c.Traces.Enabled, c.Logs.ExportOTLP = tt.traces, tt.logs
			r, err := telemetry.NewTelemetryService(t.Context(), c, telemetry.ServiceInfo{Name: "switches"})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { shutdown(t, r) })
			ctx, span := r.TracerProvider().Tracer("test").Start(t.Context(), "signal test")
			if span.IsRecording() != tt.traces {
				t.Fatal("trace switch ignored")
			}
			ctx = telemetry.WithTenantAttributes(ctx, "acme", "01")
			slog.InfoContext(ctx, "signal switches")
			span.End()
			shutdown(t, r)
			col.mu.Lock()
			defer col.mu.Unlock()
			if (len(col.spans) > 0) != tt.traces || (len(col.logs) > 0) != tt.logs {
				t.Fatalf("unexpected exports: spans=%d logs=%d", len(col.spans), len(col.logs))
			}
			if tt.logs && !strings.Contains(col.logs[0].String(), "sark.empresa") {
				t.Fatal("tenant log attributes missing")
			}
		})
	}
}

func TestTelemetryServiceParentBasedSampling(t *testing.T) {
	_, srv := newCollector(t)
	c := telemetry.DefaultConfig()
	c.Enabled, c.OTLP.Endpoint, c.Traces.SampleRatio = true, srv.URL, 0
	c.Logs.ExportOTLP = false
	r, err := telemetry.NewTelemetryService(t.Context(), c, telemetry.ServiceInfo{Name: "sampling"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdown(t, r) })
	tracer := r.TracerProvider().Tracer("test")
	_, root := tracer.Start(t.Context(), "new root")
	if root.IsRecording() {
		t.Error("explicit zero ratio ignored")
	}
	root.End()
	for _, sampled := range []bool{false, true} {
		flags := trace.TraceFlags(0)
		if sampled {
			flags = trace.FlagsSampled
		}
		sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, TraceFlags: flags, Remote: true})
		_, child := tracer.Start(trace.ContextWithRemoteSpanContext(t.Context(), sc), "child")
		if child.IsRecording() != sampled {
			t.Errorf("parent sampling decision %v ignored", sampled)
		}
		child.End()
	}
}

func TestTelemetryServiceDisabledUsesStdout(t *testing.T) {
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writer
	t.Cleanup(func() { os.Stdout = original; _ = reader.Close(); _ = writer.Close() })
	r, err := telemetry.NewTelemetryService(t.Context(), telemetry.DefaultConfig(), telemetry.ServiceInfo{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shutdown(t, r) })
	if r.Enabled() {
		t.Fatal("telemetry unexpectedly enabled")
	}
	_, span := r.TracerProvider().Tracer("test").Start(t.Context(), "noop")
	if span.IsRecording() {
		t.Error("disabled tracing is recording")
	}
	span.End()
	slog.Info("legacy slog remains visible")
	shutdown(t, r)
	_ = writer.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "legacy slog remains visible") {
		t.Fatalf("stdout missing: %s", data)
	}
}

func TestTelemetryServiceCollectorUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(503) }))
	t.Cleanup(srv.Close)
	c := telemetry.DefaultConfig()
	c.Enabled, c.OTLP.Endpoint, c.Traces.SampleRatio = true, srv.URL, 1
	r, err := telemetry.NewTelemetryService(t.Context(), c, telemetry.ServiceInfo{Name: "outage-test"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Shutdown(context.Background()) })
	_, span := r.TracerProvider().Tracer("test").Start(t.Context(), "unavailable collector")
	span.SetAttributes(attribute.String("test", "value"))
	span.End()
	slog.Info("stdout still works with Collector unavailable")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_ = r.Shutdown(ctx) // Failed export is expected, shutdown must remain bounded.
	if time.Since(start) > 2*time.Second {
		t.Fatal("Collector outage blocked shutdown")
	}
}
