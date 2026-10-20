package setup

import (
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/sfperusacdev/identitysdk/configs"
	connection "github.com/sfperusacdev/identitysdk/pg-connection"
	"github.com/sfperusacdev/identitysdk/telemetry"
	"github.com/spf13/viper"
)

func TestRunValidatesTelemetryBeforeOpeningDatabase(t *testing.T) {
	var storageCalled bool
	v := viper.New()
	v.Set("telemetry.enabled", true)
	service := NewService("test", WithConfigProvider(func(configs.ConfigPath) (configs.GeneralServiceConfigProvider, configs.DatabaseConfigProvider, *viper.Viper, error) {
		c := &configs.GeneralServiceConfig{}
		return c, c, v, nil
	}), WithStorageManagerProvider(func() (connection.StorageManager, error) {
		storageCalled = true
		return connection.SkipStorage()
	}))
	service.Command.SetArgs([]string{"--config", "custom-provider.yaml"})
	service.Command.SetOut(io.Discard)
	service.Command.SetErr(io.Discard)
	err := service.Run()
	if err == nil || !strings.Contains(err.Error(), "telemetry.otlp.endpoint") {
		t.Fatalf("expected telemetry validation error, got %v", err)
	}
	if storageCalled {
		t.Fatal("database opened before configuration validation")
	}
}

func TestRunCleansUpTelemetryOnDatabaseStartupFailure(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		name := "missing telemetry section"
		if enabled {
			name = "enabled without exporters"
		}
		t.Run(name, func(t *testing.T) {
			before, beforeLogger := telemetry.Current(), slog.Default()
			v := viper.New()
			if enabled {
				v.Set("telemetry.enabled", true)
				v.Set("telemetry.traces.enabled", false)
				v.Set("telemetry.logs.export_otlp", false)
			}
			wantErr := errors.New("database unavailable")
			service := NewService("test", WithConfigProvider(func(configs.ConfigPath) (configs.GeneralServiceConfigProvider, configs.DatabaseConfigProvider, *viper.Viper, error) {
				c := &configs.GeneralServiceConfig{}
				return c, c, v, nil
			}), WithStorageManagerProvider(func() (connection.StorageManager, error) {
				if telemetry.Current() == before || telemetry.Current().Enabled() != enabled {
					t.Error("telemetry was not initialized before database provider")
				}
				if slog.Default() == beforeLogger {
					t.Error("stdout logger was not initialized")
				}
				return nil, wantErr
			}))
			service.Command.SetArgs([]string{"--config", "custom-provider.yaml"})
			service.Command.SetOut(io.Discard)
			service.Command.SetErr(io.Discard)
			if err := service.Run(); !errors.Is(err, wantErr) {
				t.Fatalf("expected storage error, got %v", err)
			}
			if telemetry.Current() != before || slog.Default() != beforeLogger {
				t.Fatal("failed startup leaked global telemetry state")
			}
		})
	}
}
