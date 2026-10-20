package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func newTelemetryResource(ctx context.Context, name, version, environment string) (*resource.Resource, error) {
	if name == "" {
		name = "unknown_service"
	}
	attrs := []attribute.KeyValue{
		attribute.String("service.name", name),
		attribute.String("service.version", version),
		attribute.String("service.instance.id", uuid.NewString()),
	}
	if environment != "" {
		attrs = append(attrs, attribute.String("deployment.environment.name", environment))
	}
	res, err := resource.New(ctx, resource.WithFromEnv(), resource.WithAttributes(attrs...))
	if err != nil {
		return nil, fmt.Errorf("telemetry resource: %w", err)
	}
	return res, nil
}

func newTracerProvider(ctx context.Context, c Config, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	exporter, err := newOTLPTraceExporter(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("telemetry trace exporter: %w", err)
	}
	return sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(c.Traces.SampleRatio))),
		sdktrace.WithBatcher(safeSpanExporter{exporter},
			sdktrace.WithMaxQueueSize(2048), sdktrace.WithExportTimeout(5*time.Second)),
	), nil
}

func newLoggerProvider(ctx context.Context, c Config, res *resource.Resource) (*sdklog.LoggerProvider, slog.Handler, error) {
	exporter, err := newOTLPLogExporter(ctx, c)
	if err != nil {
		return nil, nil, fmt.Errorf("telemetry log exporter: %w", err)
	}
	provider := sdklog.NewLoggerProvider(sdklog.WithResource(res), sdklog.WithProcessor(
		sdklog.NewBatchProcessor(exporter, sdklog.WithMaxQueueSize(2048), sdklog.WithExportTimeout(5*time.Second)),
	))
	bridge := otelslog.NewHandler("github.com/sfperusacdev/identitysdk",
		otelslog.WithLoggerProvider(provider), otelslog.WithSource(false))
	return provider, bridge, nil
}

func newOTLPTraceExporter(ctx context.Context, c Config) (sdktrace.SpanExporter, error) {
	if c.OTLP.Protocol == "grpc" {
		return otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpointURL(c.OTLP.Endpoint), otlptracegrpc.WithTimeout(5*time.Second))
	}
	return otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(strings.TrimRight(c.OTLP.Endpoint, "/")+"/v1/traces"),
		otlptracehttp.WithTimeout(5*time.Second))
}

func newOTLPLogExporter(ctx context.Context, c Config) (sdklog.Exporter, error) {
	if c.OTLP.Protocol == "grpc" {
		return otlploggrpc.New(ctx,
			otlploggrpc.WithEndpointURL(c.OTLP.Endpoint), otlploggrpc.WithTimeout(5*time.Second))
	}
	return otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(strings.TrimRight(c.OTLP.Endpoint, "/")+"/v1/logs"),
		otlploghttp.WithTimeout(5*time.Second))
}

// SQL, errors and URLs can contain credentials; sanitize them before export.
type safeSpanExporter struct{ sdktrace.SpanExporter }

func (e safeSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	safe := make([]sdktrace.ReadOnlySpan, len(spans))
	for i, span := range spans {
		safe[i] = safeSpan{ReadOnlySpan: span}
	}
	return e.SpanExporter.ExportSpans(ctx, safe)
}

type safeSpan struct{ sdktrace.ReadOnlySpan }

func (s safeSpan) Attributes() []attribute.KeyValue {
	return safeAttributes(s.ReadOnlySpan.Attributes())
}

func (s safeSpan) Events() []sdktrace.Event {
	events := s.ReadOnlySpan.Events()
	safe := make([]sdktrace.Event, len(events))
	for i, event := range events {
		event.Attributes = safeAttributes(event.Attributes)
		safe[i] = event
	}
	return safe
}

func (s safeSpan) Status() sdktrace.Status {
	status := s.ReadOnlySpan.Status()
	status.Description = ""
	return status
}

func safeAttributes(attrs []attribute.KeyValue) []attribute.KeyValue {
	safe := make([]attribute.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		switch string(attr.Key) {
		case "url.query", "db.statement", "db.query.text", "db.connection_string", "exception.message", "exception.stacktrace", "echo.error", "error.message":
			continue
		case "url.full", "http.url":
			u, err := url.Parse(attr.Value.AsString())
			if err != nil {
				continue
			}
			u.User, u.RawQuery, u.Fragment, u.RawFragment = nil, "", "", ""
			u.ForceQuery = false
			attr = attribute.String(string(attr.Key), u.String())
		}
		safe = append(safe, attr)
	}
	return safe
}
