package telemetry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type logHandler struct {
	stdout slog.Handler
	otlp   slog.Handler
}

func (h logHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.stdout.Enabled(ctx, level) || (h.otlp != nil && h.otlp.Enabled(ctx, level))
}

func (h logHandler) Handle(ctx context.Context, r slog.Record) error {
	var stdoutErr, otlpErr error
	if attrs, ok := ctx.Value(tenantLogKey{}).([]slog.Attr); ok {
		r = r.Clone()
		r.AddAttrs(attrs...)
	}
	if h.stdout.Enabled(ctx, r.Level) {
		local := r.Clone()
		if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
			local.AddAttrs(slog.String("trace_id", sc.TraceID().String()), slog.String("span_id", sc.SpanID().String()))
		}
		stdoutErr = h.stdout.Handle(ctx, local)
	}
	if h.otlp != nil && h.otlp.Enabled(ctx, r.Level) {
		otlpErr = h.otlp.Handle(ctx, r.Clone())
	}
	return errors.Join(stdoutErr, otlpErr)
}

func (h logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.stdout = h.stdout.WithAttrs(attrs)
	if h.otlp != nil {
		h.otlp = h.otlp.WithAttrs(attrs)
	}
	return h
}

func (h logHandler) WithGroup(name string) slog.Handler {
	h.stdout = h.stdout.WithGroup(name)
	if h.otlp != nil {
		h.otlp = h.otlp.WithGroup(name)
	}
	return h
}

func newTelemetryLogger(c Config, stdout io.Writer, otlp slog.Handler) *slog.Logger {
	level := slog.LevelInfo
	_ = level.UnmarshalText([]byte(c.Logs.Level))
	opts := &slog.HandlerOptions{Level: level}
	var h slog.Handler = slog.NewTextHandler(stdout, opts)
	if c.Enabled && c.Logs.Format == "json" {
		h = slog.NewJSONHandler(stdout, opts)
	}
	// The OTLP bridge does not enforce the application's log level.
	if otlp != nil {
		otlp = levelHandler{Handler: otlp, level: level}
	}
	return slog.New(logHandler{stdout: h, otlp: otlp})
}

type levelHandler struct {
	slog.Handler
	level slog.Level
}

func (h levelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level && h.Handler.Enabled(ctx, level)
}

func (h levelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.Handler = h.Handler.WithAttrs(attrs)
	return h
}

func (h levelHandler) WithGroup(name string) slog.Handler {
	h.Handler = h.Handler.WithGroup(name)
	return h
}

type tenantLogKey struct{}

// Only use tenant values from an already validated context.
func WithTenantAttributes(ctx context.Context, empresa, sucursal string) context.Context {
	if !Current().Enabled() {
		return ctx
	}
	var attrs []slog.Attr
	for _, value := range []struct{ key, value string }{{"sark.empresa", empresa}, {"sark.sucursal", sucursal}} {
		if value.value != "" && !strings.Contains(value.value, "####") {
			trace.SpanFromContext(ctx).SetAttributes(attribute.String(value.key, value.value))
			attrs = append(attrs, slog.String(value.key, value.value))
		}
	}
	if len(attrs) == 0 {
		return ctx
	}
	return context.WithValue(ctx, tenantLogKey{}, attrs)
}

// Never report exporter failures through the failing OTLP handler.
func newExportErrorHandler(stdout io.Writer) otel.ErrorHandler {
	diagnostic := slog.New(slog.NewTextHandler(stdout, nil))
	var lastWarning atomic.Int64
	return otel.ErrorHandlerFunc(func(err error) {
		now := time.Now().Unix()
		last := lastWarning.Load()
		if now-last >= 60 && lastWarning.CompareAndSwap(last, now) {
			diagnostic.Warn("telemetry export failed", "error_type", fmt.Sprintf("%T", err))
		}
	})
}
