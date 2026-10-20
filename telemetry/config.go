package telemetry

import (
	"fmt"
	"math"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Enabled     bool   `mapstructure:"enabled"`
	Environment string `mapstructure:"environment"`
	OTLP        struct {
		Endpoint string `mapstructure:"endpoint"`
		Protocol string `mapstructure:"protocol"`
	} `mapstructure:"otlp"`
	Traces struct {
		Enabled     bool    `mapstructure:"enabled"`
		SampleRatio float64 `mapstructure:"sample_ratio"`
	} `mapstructure:"traces"`
	Logs struct {
		Level      string `mapstructure:"level"`
		Format     string `mapstructure:"format"`
		ExportOTLP bool   `mapstructure:"export_otlp"`
	} `mapstructure:"logs"`
}

func DefaultConfig() Config {
	var c Config
	c.OTLP.Protocol = "http/protobuf"
	c.Traces.Enabled = true
	c.Traces.SampleRatio = 0.1
	c.Logs.Level = "info"
	c.Logs.Format = "json"
	c.Logs.ExportOTLP = true
	return c
}

func Read(v *viper.Viper) (Config, error) {
	// Decode over defaults so explicit false and zero are preserved.
	c := DefaultConfig()
	if v != nil {
		if err := v.UnmarshalKey("telemetry", &c); err != nil {
			return c, fmt.Errorf("decode telemetry: %w", err)
		}
	}
	c.Environment = strings.TrimSpace(c.Environment)
	c.OTLP.Endpoint = strings.TrimSpace(c.OTLP.Endpoint)
	return c, c.Validate()
}

func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.OTLP.Protocol != "http/protobuf" && c.OTLP.Protocol != "grpc" {
		return fmt.Errorf("telemetry.otlp.protocol must be http/protobuf or grpc")
	}
	if math.IsNaN(c.Traces.SampleRatio) || math.IsInf(c.Traces.SampleRatio, 0) || c.Traces.SampleRatio < 0 || c.Traces.SampleRatio > 1 {
		return fmt.Errorf("telemetry.traces.sample_ratio must be between 0 and 1")
	}
	switch c.Logs.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("telemetry.logs.level must be debug, info, warn or error")
	}
	if c.Logs.Format != "json" && c.Logs.Format != "text" {
		return fmt.Errorf("telemetry.logs.format must be json or text")
	}
	if c.Traces.Enabled || c.Logs.ExportOTLP {
		u, err := url.Parse(c.OTLP.Endpoint)
		if err != nil || u == nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return fmt.Errorf("telemetry.otlp.endpoint must be an http(s) URL when exporting signals")
		}
		if u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			return fmt.Errorf("telemetry.otlp.endpoint must not contain credentials, query parameters or a fragment")
		}
		if c.OTLP.Protocol == "grpc" && u.Path != "" && u.Path != "/" {
			return fmt.Errorf("telemetry.otlp.endpoint must not contain a path for grpc")
		}
	}
	return nil
}
