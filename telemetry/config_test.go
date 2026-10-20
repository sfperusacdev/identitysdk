package telemetry

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestRead(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		check   func(*testing.T, Config)
	}{
		{name: "missing section", yaml: "address: ':8080'", check: func(t *testing.T, c Config) {
			if c.Enabled || c.OTLP.Protocol != "http/protobuf" || !c.Traces.Enabled || c.Traces.SampleRatio != 0.1 || !c.Logs.ExportOTLP || c.Logs.Level != "info" || c.Logs.Format != "json" || c.Environment != "" {
				t.Fatalf("unexpected defaults: %+v", c)
			}
		}},
		{name: "minimal enabled", yaml: "telemetry:\n  enabled: true\n  otlp:\n    endpoint: http://localhost:4318", check: func(t *testing.T, c Config) {
			if !c.Enabled || !c.Traces.Enabled || !c.Logs.ExportOTLP {
				t.Fatalf("missing enabled defaults: %+v", c)
			}
		}},
		{name: "explicit zero and false", yaml: "telemetry:\n  enabled: true\n  traces:\n    enabled: false\n    sample_ratio: 0\n  logs:\n    export_otlp: false", check: func(t *testing.T, c Config) {
			if c.Traces.Enabled || c.Traces.SampleRatio != 0 || c.Logs.ExportOTLP {
				t.Fatalf("explicit values overwritten: %+v", c)
			}
		}},
		{name: "disabled ignores inactive settings", yaml: "telemetry:\n  enabled: false\n  logs:\n    level: unused"},
		{name: "missing endpoint", yaml: "telemetry:\n  enabled: true", wantErr: true},
		{name: "unsupported protocol", yaml: "telemetry:\n  enabled: true\n  otlp:\n    protocol: udp", wantErr: true},
		{name: "invalid ratio", yaml: "telemetry:\n  enabled: true\n  traces:\n    sample_ratio: 1.1", wantErr: true},
		{name: "invalid level", yaml: "telemetry:\n  enabled: true\n  logs:\n    level: trace", wantErr: true},
		{name: "invalid format", yaml: "telemetry:\n  enabled: true\n  logs:\n    format: xml", wantErr: true},
		{name: "invalid type", yaml: "telemetry:\n  enabled: not-a-bool", wantErr: true},
		{name: "credentials forbidden", yaml: "telemetry:\n  enabled: true\n  otlp:\n    endpoint: https://user:secret@collector:4318", wantErr: true},
		{name: "query forbidden", yaml: "telemetry:\n  enabled: true\n  otlp:\n    endpoint: https://collector:4318?token=secret", wantErr: true},
		{name: "grpc endpoint", yaml: "telemetry:\n  enabled: true\n  otlp:\n    endpoint: http://collector:4317\n    protocol: grpc"},
		{name: "grpc path forbidden", yaml: "telemetry:\n  enabled: true\n  otlp:\n    endpoint: http://collector:4317/path\n    protocol: grpc", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := viper.New()
			v.SetConfigType("yaml")
			if err := v.ReadConfig(strings.NewReader(tt.yaml)); err != nil {
				t.Fatal(err)
			}
			c, err := Read(v)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Read error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.check != nil {
				tt.check(t, c)
			}
		})
	}
	if c, err := Read(nil); err != nil || c.Enabled {
		t.Fatalf("nil Viper should disable telemetry: %+v, %v", c, err)
	}
}

func TestReadExampleConfig(t *testing.T) {
	t.Parallel()
	v := viper.New()
	v.SetConfigFile("config.example.yml")
	if err := v.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	c, err := Read(v)
	if err != nil {
		t.Fatal(err)
	}
	want := DefaultConfig()
	want.Enabled = true
	want.OTLP.Endpoint = "http://localhost:4318"
	if c != want {
		t.Fatalf("unexpected example configuration: %+v", c)
	}
}
