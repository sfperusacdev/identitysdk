package identitysdk

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

func IdentityServerCheckHealth() error {
	return IdentityServerCheckHealthWithContext(context.Background())
}

func IdentityServerCheckHealthWithContext(ctx context.Context) error {
	healthURL, err := url.JoinPath(identityAddress, "/health")
	if err != nil {
		slog.Error("failed to construct health check URL", "error", err)
		return fmt.Errorf("failed to construct health check URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		slog.Error("failed to create HTTP request", "url", healthURL, "error", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("failed to send health check request", "url", healthURL, "error", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(res.Body)
		slog.Error("health check failed",
			"url", healthURL,
			"status", res.StatusCode,
			"response", string(payload),
		)
		return fmt.Errorf("identity server error: %s", string(payload))
	}
	return nil
}
