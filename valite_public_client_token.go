package identitysdk

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/sfperusacdev/identitysdk/entities"
	"github.com/user0608/goones/errs"
)

func ValidatePublicClientToken(ctx context.Context, token string) (*entities.JwtPublicClientData, error) {
	hostURL, err := url.JoinPath(identityAddress, "/v1/check-public-client-token")
	if err != nil {
		slog.Error("failed to construct client token validation URL", "error", err)
		return nil, errs.InternalErrorDirect(errs.ErrInternal)
	}

	var buff bytes.Buffer
	payload := struct {
		Token string `json:"token"`
	}{Token: token}

	if err := json.NewEncoder(&buff).Encode(&payload); err != nil {
		slog.Error("failed to encode client token payload", "error", err)
		return nil, errs.InternalErrorDirect(errs.ErrInternal)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hostURL, &buff)
	if err != nil {
		slog.Error("failed to create HTTP request", "url", hostURL, "error", err)
		return nil, errs.InternalErrorDirect(errs.ErrInternal)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("failed to send token validation request", "url", hostURL, "error", err)
		return nil, errs.InternalErrorDirect("Auth server no responde")
	}
	defer res.Body.Close()

	var response struct {
		Type    string                       `json:"type"`
		Message string                       `json:"message"`
		Data    entities.JwtPublicClientData `json:"data"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		slog.Error("failed to decode client token validation response", "url", hostURL, "error", err)
		return nil, errs.InternalErrorDirect(errs.ErrInternal)
	}

	if res.StatusCode != http.StatusOK {
		slog.Warn("token validation failed", "url", hostURL, "status", res.StatusCode, "message", response.Message)
		return nil, errs.BadRequestDirect(response.Message)
	}

	return &response.Data, nil
}
