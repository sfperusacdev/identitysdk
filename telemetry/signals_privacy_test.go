package telemetry

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestSafeSpanExporter(t *testing.T) {
	t.Parallel()
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(safeSpanExporter{exporter}))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	_, span := provider.Tracer("test").Start(t.Context(), "operation")
	span.SetAttributes(
		attribute.String("url.full", "https://user:URL-SECRET@example.com/path?token=QUERY-SECRET#FRAGMENT-SECRET"),
		attribute.String("url.query", "token=QUERY-SECRET"),
		attribute.String("db.query.text", "SELECT 'SQL-SECRET'"),
		attribute.String("db.connection_string", "password=DB-SECRET"),
		attribute.String("http.route", "/workers/:code"),
	)
	span.RecordError(errors.New("ERROR-SECRET"))
	span.SetStatus(codes.Error, "STATUS-SECRET")
	span.End()
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("exported %d spans", len(spans))
	}
	if spans[0].Status.Code != codes.Error || spans[0].Status.Description != "" {
		t.Fatal("error status not safely retained")
	}
	for _, attr := range spans[0].Attributes {
		if strings.Contains(attr.Value.Emit(), "SECRET") {
			t.Errorf("secret in attribute %s", attr.Key)
		}
		if attr.Key == "url.full" && attr.Value.AsString() != "https://example.com/path" {
			t.Errorf("URL changed unexpectedly: %v", attr)
		}
	}
	for _, event := range spans[0].Events {
		for _, attr := range event.Attributes {
			if strings.Contains(attr.Value.Emit(), "SECRET") {
				t.Errorf("secret in event %s", attr.Key)
			}
		}
	}
}
