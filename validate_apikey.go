package identitysdk

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/sfperusacdev/identitysdk/entities"
	apikeycache "github.com/sfperusacdev/identitysdk/internal/apikey_cache"
	"github.com/user0608/goones/errs"
	"github.com/user0608/ifdevmode"
)

func ValidateApiKeyWithCache(ctx context.Context, apikey string) (*entities.ApikeyData, error) {
	var cacheData = apikeycache.DefaultCache.Get(ctx, apikey)
	if cacheData != nil {
		if ifdevmode.Yes() {
			slog.Info("Session api key data read from cache",
				"empresa", cacheData.Apikey.Empresa,
			)
		}
		return cacheData, nil
	}
	apikeydata, err := ValidateApiKey(ctx, apikey)
	if err != nil {
		return nil, err
	}
	if apikeydata != nil {
		apikeycache.DefaultCache.Set(ctx, apikey, *apikeydata)
	}
	return apikeydata, nil
}

func ValidateApiKey(ctx context.Context, apikey string) (*entities.ApikeyData, error) {
	hostURL, err := url.JoinPath(identityAddress, "/v1/check-apikey")
	if err != nil {
		slog.Error("failed to construct API key validation URL", "error", err)
		return nil, errs.InternalErrorDirect(errs.ErrInternal)
	}

	var buff bytes.Buffer
	payload := struct {
		Apikey string `json:"apikey"`
	}{Apikey: apikey}

	if err := json.NewEncoder(&buff).Encode(&payload); err != nil {
		slog.Error("failed to encode API key payload", "error", err)
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
		slog.Error("failed to send API key validation request", "url", hostURL, "error", err)
		return nil, errs.InternalErrorDirect("Auth server no responde")
	}
	defer res.Body.Close()

	var response struct {
		Type    string              `json:"type"`
		Message string              `json:"message"`
		Data    entities.ApikeyData `json:"data"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		slog.Error("failed to decode API key validation response", "url", hostURL, "error", err)
		return nil, errs.InternalErrorDirect(errs.ErrInternal)
	}

	if res.StatusCode != http.StatusOK {
		slog.Warn("API key validation failed", "url", hostURL, "status", res.StatusCode, "message", response.Message)
		return nil, errs.BadRequestDirect(response.Message)
	}

	return &response.Data, nil
}
