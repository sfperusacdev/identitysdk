package telemetry

import (
	"testing"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

func TestNewTelemetryResource(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, serviceName, environment, wantName string }{
		{"configured service", "contratos_api", "production", "contratos_api"},
		{"missing service name", "", "", "unknown_service"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, err := newTelemetryResource(t.Context(), tt.serviceName, "1.2.3", tt.environment)
			if err != nil {
				t.Fatal(err)
			}
			attrs := map[attribute.Key]string{}
			for _, attr := range res.Attributes() {
				attrs[attr.Key] = attr.Value.AsString()
			}
			if attrs["service.name"] != tt.wantName || attrs["service.version"] != "1.2.3" {
				t.Fatalf("unexpected identity: %v", attrs)
			}
			if tt.environment != "" && attrs["deployment.environment.name"] != tt.environment {
				t.Fatal("deployment environment was lost")
			}
			if _, err := uuid.Parse(attrs["service.instance.id"]); err != nil {
				t.Fatalf("invalid instance identity: %v", err)
			}
		})
	}
}
