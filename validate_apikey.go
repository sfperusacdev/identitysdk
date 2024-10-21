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

func ValidateApiKey(ctx context.Context, apikey string) (data *entities.ApikeyData, err error) {
	hostUrl, err := url.JoinPath(identityAddress, "/v1/check-apikey")
	if err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		return nil, errs.Internal(errs.ErrInternal)
	}
	var buff bytes.Buffer
	var payload = struct {
		Apikey string `json:"apikey"`
	}{Apikey: apikey}
	if err := json.NewEncoder(&buff).Encode(&payload); err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		return nil, errs.Internal(errs.ErrInternal)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hostUrl, &buff)
	if err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		return nil, errs.Internal(errs.ErrInternal)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		return nil, errs.Internal("Auth server no responde")
	}
	defer res.Body.Close()
	var response struct {
		Type    string              `json:"type"`
		Message string              `json:"message"`
		Data    entities.ApikeyData `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		return nil, errs.Internal(errs.ErrInternal)
	}
	if res.StatusCode != http.StatusOK {
		return nil, errs.Bad(response.Message)
	}
	return &response.Data, nil
}
