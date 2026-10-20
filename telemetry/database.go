package telemetry

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

func (s *TelemetryService) InstrumentDB(db *gorm.DB) error {
	if !s.Enabled() || !s.config.Traces.Enabled || db == nil {
		return nil
	}
	if _, loaded := s.instrumentedDBs.LoadOrStore(db.Config, struct{}{}); loaded {
		return nil
	}
	if _, exists := db.Config.Plugins["otelgorm"]; exists {
		return nil
	}
	if err := db.Use(tracing.NewPlugin(
		tracing.WithTracerProvider(s.tracerProvider), tracing.WithoutMetrics(),
		tracing.WithoutQueryVariables(), tracing.WithoutServerAddress(),
		// Exclude literals embedded directly in raw SQL as well.
		tracing.WithQueryFormatter(sqlOperation),
	)); err != nil {
		return fmt.Errorf("instrument GORM: %w", err)
	}
	return nil
}

func sqlOperation(query string) string {
	fields := strings.Fields(query)
	if len(fields) > 0 {
		switch op := strings.ToUpper(fields[0]); op {
		case "SELECT", "INSERT", "UPDATE", "DELETE", "WITH", "CREATE", "ALTER", "DROP", "BEGIN", "COMMIT", "ROLLBACK":
			return op
		}
	}
	return "SQL"
}
