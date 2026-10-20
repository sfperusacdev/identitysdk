package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

type logExporter struct{ records []sdklog.Record }

func (e *logExporter) Export(_ context.Context, records []sdklog.Record) error {
	for _, r := range records {
		e.records = append(e.records, r.Clone())
	}
	return nil
}
func (*logExporter) Shutdown(context.Context) error   { return nil }
func (*logExporter) ForceFlush(context.Context) error { return nil }

func TestNewTelemetryLoggerStdoutAndOTLP(t *testing.T) {
	t.Parallel()
	exporter := &logExporter{}
	provider := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewSimpleProcessor(exporter)))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	bridge := otelslog.NewHandler("test", otelslog.WithLoggerProvider(provider))
	var stdout bytes.Buffer
	c := DefaultConfig()
	c.Enabled = true
	logger := newTelemetryLogger(c, &stdout, bridge)
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)
	logger.With("component", "test").WithGroup("operation").DebugContext(ctx, "hidden")
	logger.With("component", "test").WithGroup("operation").InfoContext(ctx, "existing slog", "code", "123")
	if len(exporter.records) != 1 {
		t.Fatalf("exported %d records, want 1", len(exporter.records))
	}
	r := exporter.records[0]
	if r.TraceID() != sc.TraceID() || r.SpanID() != sc.SpanID() || r.Body().AsString() != "existing slog" {
		t.Fatalf("missing OTLP correlation: %+v", r)
	}
	var local map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &local); err != nil {
		t.Fatal(err)
	}
	if local["msg"] != "existing slog" || local["component"] != "test" {
		t.Fatalf("stdout lost existing fields: %s", stdout.String())
	}
	operation := local["operation"].(map[string]any)
	if operation["trace_id"] != sc.TraceID().String() || operation["span_id"] != sc.SpanID().String() || operation["code"] != "123" {
		t.Fatalf("stdout lost context/group fields: %s", stdout.String())
	}
}

func TestNewTelemetryLoggerStdoutOnly(t *testing.T) {
	t.Parallel()
	for _, enabled := range []bool{false, true} {
		name := "disabled"
		if enabled {
			name = "enabled without log export"
		}
		t.Run(name, func(t *testing.T) {
			var stdout bytes.Buffer
			c := DefaultConfig()
			c.Enabled, c.Logs.Format = enabled, "text"
			logger := newTelemetryLogger(c, &stdout, nil)
			logger.Log(context.Background(), slog.LevelInfo, "visible on stdout")
			if !strings.Contains(stdout.String(), "visible on stdout") {
				t.Fatalf("stdout is missing: %q", stdout.String())
			}
		})
	}
}
