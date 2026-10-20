package telemetry

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestHTTPClientPreservesPolicyAndPropagatesTrace(t *testing.T) {
	t.Parallel()
	provider := sdktrace.NewTracerProvider()
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	ctx, parent := provider.Tracer("test").Start(t.Context(), "parent")
	defer parent.End()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	var called bool
	original := &http.Client{
		Timeout:       time.Second,
		Jar:           jar,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			called = true
			sc := trace.SpanContextFromContext(propagation.TraceContext{}.Extract(req.Context(), propagation.HeaderCarrier(req.Header)))
			if sc.TraceID() != parent.SpanContext().TraceID() {
				t.Error("trace was not propagated")
			}
			if req.Header.Get("Authorization") != "test-token" {
				t.Error("application headers changed")
			}
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Header: make(http.Header)}, nil
		}),
	}
	service := &TelemetryService{tracerProvider: provider, propagator: propagation.TraceContext{}}
	client := service.httpClient(original)
	if client == original || client.Timeout != original.Timeout || client.Jar != jar || client.CheckRedirect == nil {
		t.Fatal("client was not copied with its existing policy")
	}
	if _, ok := original.Transport.(roundTripFunc); !ok {
		t.Fatal("original transport was mutated")
	}
	if err := client.CheckRedirect(nil, nil); err != http.ErrUseLastResponse {
		t.Fatalf("redirect policy changed: %v", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.test/path", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "test-token")
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if !called || req.Header.Get("traceparent") != "" {
		t.Fatal("transport was not called or the original request was mutated")
	}
}
