package telemetry_test

import (
	"bytes"
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk/telemetry"
	"github.com/sfperusacdev/identitysdk/xreq"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
)

// No parallelism: TelemetryService owns process-wide slog and OTel defaults.
func TestTelemetryServiceHTTPAndDatabase(t *testing.T) {
	col, srv := newCollector(t)
	r := startTelemetryService(t, srv.URL+"/")
	pool, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: pool}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.InstrumentDB(db); err != nil {
		t.Fatal(err)
	}
	if err := r.InstrumentDB(db); err != nil {
		t.Fatal(err)
	}
	backend := echo.New()
	backend.Use(r.HTTPMiddleware(), telemetry.AccessLogger)
	backend.GET("/workers/:code", func(c echo.Context) error {
		if c.QueryParam("token") != "URL-SECRET" || c.Request().Header.Get("Authorization") != "AUTH-SECRET" {
			return echo.NewHTTPError(400, "credentials lost")
		}
		var result string
		if err := db.WithContext(c.Request().Context()).Raw("SELECT ?", "SQL-SECRET").Scan(&result).Error; err != nil {
			return err
		}
		if result != "SQL-SECRET" {
			return echo.NewHTTPError(500, "query changed")
		}
		slog.InfoContext(c.Request().Context(), "backend completed")
		return c.JSON(200, map[string]bool{"ok": true})
	})
	backendServer := httptest.NewServer(backend)
	t.Cleanup(backendServer.Close)
	frontend := echo.New()
	frontend.Use(r.HTTPMiddleware(), telemetry.AccessLogger)
	frontend.GET("/api/:code", func(c echo.Context) error {
		err := xreq.MakeRequest(c.Request().Context(), backendServer.URL, "/workers/123",
			xreq.WithAuthorization("AUTH-SECRET"), xreq.WithQueryParam("token", "URL-SECRET"))
		if err != nil {
			return err
		}
		return c.NoContent(204)
	})
	ctx, parent := r.TracerProvider().Tracer("test").Start(t.Context(), "upstream")
	req := httptest.NewRequest(http.MethodGet, "/api/123", nil).WithContext(ctx)
	response := httptest.NewRecorder()
	frontend.ServeHTTP(response, req)
	parent.End()
	if response.Code != 204 {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	shutdown(t, r)
	col.mu.Lock()
	defer col.mu.Unlock()
	var serverSpans, clientSpans, dbSpans int
	for _, span := range col.spans {
		wantTrace := parent.SpanContext().TraceID()
		if !bytes.Equal(span.TraceId, wantTrace[:]) {
			t.Errorf("trace discontinuity: %s", span.Name)
		}
		if span.Kind == tracepb.Span_SPAN_KIND_SERVER {
			serverSpans++
		}
		if span.Kind == tracepb.Span_SPAN_KIND_CLIENT {
			clientSpans++
		}
		for _, attr := range span.Attributes {
			if string(attr.Key) == "db.system.name" {
				dbSpans++
			}
		}
		for _, secret := range []string{"URL-SECRET", "AUTH-SECRET", "SQL-SECRET", "db.query.text", "url.query"} {
			if strings.Contains(span.String(), secret) {
				t.Errorf("sensitive span data: %s", span.String())
			}
		}
	}
	if serverSpans != 2 || clientSpans < 2 || dbSpans != 1 {
		t.Fatalf("unexpected coverage servers=%d clients=%d db=%d", serverSpans, clientSpans, dbSpans)
	}
	var correlated bool
	for _, log := range col.logs {
		if log.Body.GetStringValue() == "backend completed" {
			wantTrace := parent.SpanContext().TraceID()
			correlated = bytes.Equal(log.TraceId, wantTrace[:]) && len(log.SpanId) == 8
		}
	}
	if !correlated {
		t.Error("missing correlated slog export")
	}
	resources := strings.Join(col.resources, " ")
	for _, want := range []string{"test-service", "1.2.3", "deployment.environment.name", "test"} {
		if !strings.Contains(resources, want) {
			t.Errorf("missing resource %s", want)
		}
	}
}
