package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
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

func TestNewExportErrorHandler(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		err    error
		want   []string
		secret string
	}{
		{
			name: "unauthorized logs",
			err: errors.Join(&signalExportError{signal: "logs", err: fmt.Errorf(
				"failed to send logs to http://collector/api/default/v1/logs: 401 Unauthorized (body: token=secret)")}),
			want: []string{"signal=logs", "reason=http_status", "http_status=401"}, secret: "secret",
		},
		{
			name: "missing trace route",
			err: &signalExportError{signal: "traces", err: fmt.Errorf(
				"failed to send to http://collector/api/default/v1/traces: 404 Not Found (body: secret)")},
			want: []string{"signal=traces", "reason=http_status", "http_status=404"}, secret: "secret",
		},
		{
			name: "connection failure",
			err: &signalExportError{signal: "logs", err: &url.Error{
				Op: "Post", URL: "https://secret@collector/v1/logs", Err: errors.New("connection refused"),
			}},
			want: []string{"signal=logs", "reason=network"}, secret: "secret",
		},
		{
			name: "timeout",
			err:  &signalExportError{signal: "traces", err: context.DeadlineExceeded},
			want: []string{"signal=traces", "reason=timeout"},
		},
		{
			name: "unknown failure",
			err:  errors.New("secret details"),
			want: []string{"reason=other"}, secret: "secret details",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			newExportErrorHandler(&output).Handle(tc.err)
			for _, want := range tc.want {
				if !strings.Contains(output.String(), want) {
					t.Errorf("missing %q in diagnostic: %s", want, output.String())
				}
			}
			if tc.secret != "" && strings.Contains(output.String(), tc.secret) {
				t.Errorf("diagnostic leaked secret: %s", output.String())
			}
		})
	}
}
