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

func validateTokenWithCache(ctx context.Context, token string) (*entities.JwtData, error) {
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

func ValidateToken(ctx context.Context, token string) (data *entities.JwtData, err error) {
	hostUrl, err := url.JoinPath(identityAddress, "/v1/check-token")
	if err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		return nil, errs.Internal(errs.ErrInternal)
	}
	var buff bytes.Buffer
	var payload = struct {
		Token string `json:"token"`
	}{Token: token}
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
		Type    string           `json:"type"`
		Message string           `json:"message"`
		Data    entities.JwtData `json:"data"`
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
