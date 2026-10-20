package identitysdk_test

import (
	"context"
	"testing"
	"time"

	"github.com/sfperusacdev/identitysdk"
	"go.opentelemetry.io/otel/trace"
)

func TestCloneContextPreservesTraceWithoutRequestLifetime(t *testing.T) {
	t.Parallel()
	type requestKey struct{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	ctx = context.WithValue(ctx, requestKey{}, "do not retain")
	ctx = identitysdk.CtxWithDomain(ctx, "acme")
	ctx = identitysdk.CtxWithSucursal(ctx, "01")
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, TraceFlags: trace.FlagsSampled})
	ctx = trace.ContextWithSpanContext(ctx, sc)
	cancel()
	clone := identitysdk.CloneContext(ctx)
	if clone.Err() != nil || clone.Done() != nil {
		t.Fatal("clone retained cancellation")
	}
	if _, hasDeadline := clone.Deadline(); hasDeadline {
		t.Fatal("clone retained deadline")
	}
	if clone.Value(requestKey{}) != nil {
		t.Fatal("clone retained arbitrary request values")
	}
	if !trace.SpanContextFromContext(clone).Equal(sc) {
		t.Fatal("clone lost trace identity")
	}
	empresa, sucursal := identitysdk.Empresa_Sucursal(clone)
	if empresa != "acme" || sucursal != "01" {
		t.Fatalf("clone lost tenant: %s/%s", empresa, sucursal)
	}
}

func TestFireAndForgetPreservesTrace(t *testing.T) {
	t.Parallel()
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{3}, SpanID: trace.SpanID{4}})
	ctx, cancel := context.WithCancel(trace.ContextWithSpanContext(t.Context(), sc))
	cancel()
	result := make(chan context.Context, 1)
	identitysdk.FireAndForget(ctx, func(ctx context.Context) { result <- ctx })
	select {
	case clone := <-result:
		if clone.Err() != nil || !trace.SpanContextFromContext(clone).Equal(sc) {
			t.Fatal("background correlation or detachment lost")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("background work did not complete")
	}
}
