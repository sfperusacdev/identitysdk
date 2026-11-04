package telemetry

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Enabled     bool   `mapstructure:"enabled"`
	Environment string `mapstructure:"environment"`
	OTLP        struct {
		Endpoint string            `mapstructure:"endpoint"`
		Protocol string            `mapstructure:"protocol"`
		Headers  map[string]string `mapstructure:"headers"`
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

var headerEnvReference = regexp.MustCompile(`^\$\{([A-Za-z_][A-Za-z0-9_]*)\}$`)

func (c Config) resolveHeaders() (Config, error) {
	if !c.Enabled || (!c.Traces.Enabled && !c.Logs.ExportOTLP) || len(c.OTLP.Headers) == 0 {
		return c, nil
	}
	resolved := make(map[string]string, len(c.OTLP.Headers))
	for name, value := range c.OTLP.Headers {
		if strings.TrimSpace(name) == "" || strings.ContainsAny(name, "\r\n") {
			return c, fmt.Errorf("telemetry.otlp.headers contains an invalid header name")
		}
		if match := headerEnvReference.FindStringSubmatch(value); match != nil {
			var present bool
			value, present = os.LookupEnv(match[1])
			if !present || value == "" {
				return c, fmt.Errorf("telemetry.otlp.headers.%s requires environment variable %s", name, match[1])
			}
		} else if strings.Contains(value, "${") {
			return c, fmt.Errorf("telemetry.otlp.headers.%s must use a complete ${VARIABLE} reference", name)
		}
		if value == "" || strings.ContainsAny(value, "\r\n") {
			return c, fmt.Errorf("telemetry.otlp.headers.%s has an invalid value", name)
		}
		resolved[name] = value
	}
	c.OTLP.Headers = resolved
	return c, nil
}

func (c Config) withDefaultStreamName(serviceID string) Config {
	// Preserve the exporters' standard environment-based headers when no YAML
	// headers are configured.
	if !c.Enabled || (!c.Traces.Enabled && !c.Logs.ExportOTLP) || len(c.OTLP.Headers) == 0 {
		return c
	}
	for name := range c.OTLP.Headers {
		if strings.EqualFold(name, "stream-name") {
			return c
		}
	}
	if serviceID == "" {
		serviceID = "unknown_service"
	}
	// OpenObserve formats stream names; keep service.name unchanged in the
	// exported resource, but use a predictable stream-safe spelling here.
	streamName := strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + ('a' - 'A')
		}
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' {
			return r
		}
		return '_'
	}, serviceID)
	if strings.Trim(streamName, "_") == "" {
		streamName = "unknown_service"
	}
	// The resolved map belongs to this service; do not mutate the input map
	// in case a caller reuses the Config.
	headers := make(map[string]string, len(c.OTLP.Headers)+1)
	for name, value := range c.OTLP.Headers {
		headers[name] = value
	}
	headers["stream-name"] = streamName
	c.OTLP.Headers = headers
	return c
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
