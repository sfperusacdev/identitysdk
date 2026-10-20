package telemetry

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	echootel "github.com/labstack/echo-otel/v4"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
)

var noopMeterProvider = metricnoop.NewMeterProvider()

func DoHTTP(req *http.Request) (*http.Response, error) {
	service := Current()
	if !service.Enabled() {
		return http.DefaultClient.Do(req)
	}
	return service.httpClient(http.DefaultClient).Do(req)
}

func (s *TelemetryService) HTTPMiddleware() echo.MiddlewareFunc {
	return echootel.NewMiddlewareWithConfig(echootel.Config{
		TracerProvider: s.tracerProvider,
		MeterProvider:  noopMeterProvider,
		Propagators:    s.propagator,
	})
}

func (s *TelemetryService) httpClient(client *http.Client) *http.Client {
	copy := *client
	copy.Transport = otelhttp.NewTransport(copy.Transport,
		otelhttp.WithTracerProvider(s.tracerProvider),
		otelhttp.WithMeterProvider(noopMeterProvider),
		otelhttp.WithPropagators(s.propagator),
	)
	return &copy
}

func AccessLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		if err != nil {
			c.Error(err)
		}
		if !strings.Contains(c.Path(), "/api/v1/_/system_properties") {
			route := c.Path()
			if route == "" {
				route = "unmatched"
			}
			ctx := c.Request().Context()
			level := slog.LevelInfo
			if c.Response().Status >= 500 {
				level = slog.LevelError
			}
			slog.Log(ctx, level, "HTTP request", "method", c.Request().Method,
				"route", route, "status", c.Response().Status,
				"duration", time.Since(start), "response_bytes", c.Response().Size)
		}
		// c.Error already handled the response.
		return nil
	}
}
