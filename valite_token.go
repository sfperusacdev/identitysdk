package identitysdk

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/sfperusacdev/identitysdk/entities"
	sessioncache "github.com/sfperusacdev/identitysdk/internal/session_cache"
	"github.com/user0608/goones/errs"
	"github.com/user0608/ifdevmode"
)

func ValidateTokenWithCache(ctx context.Context, token string) (*entities.JwtData, error) {
	var cacheData = sessioncache.DefaultCache.Get(ctx, token)
	if cacheData != nil {
		if ifdevmode.Yes() {
			slog.Info("Session data read from cache",
				"empresa", cacheData.Jwt.Empresa,
				"usuario", cacheData.Jwt.Username,
				"trabajador", cacheData.Jwt.TabajadorCodigo,
				"UsuarioReff", cacheData.Jwt.UsuarioReff,
			)
		}
		return cacheData, nil
	}
	jwtData, err := ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if jwtData != nil {
		sessioncache.DefaultCache.Set(ctx, token, *jwtData)
	}
	return jwtData, nil
}

func ValidateToken(ctx context.Context, token string) (*entities.JwtData, error) {
	hostURL, err := url.JoinPath(identityAddress, "/v1/check-token")
	if err != nil {
		slog.Error("failed to construct token validation URL", "error", err)
		return nil, errs.InternalErrorDirect(errs.ErrInternal)
	}

	var buff bytes.Buffer
	payload := struct {
		Token string `json:"token"`
	}{Token: token}

	if err := json.NewEncoder(&buff).Encode(&payload); err != nil {
		slog.Error("failed to encode token payload", "error", err)
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
		Type    string           `json:"type"`
		Message string           `json:"message"`
		Data    entities.JwtData `json:"data"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		slog.Error("failed to decode token validation response", "url", hostURL, "error", err)
		return nil, errs.InternalErrorDirect(errs.ErrInternal)
	}

	if res.StatusCode != http.StatusOK {
		slog.Warn("token validation failed", "url", hostURL, "status", res.StatusCode, "message", response.Message)
		return nil, errs.BadRequestDirect(response.Message)
	}

	return &response.Data, nil
}
