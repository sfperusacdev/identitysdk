package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/sfperusacdev/identitysdk"
)

var ErrNotFound = errors.New("the record you are looking for was not found")

type ExternalBridgeService struct{}

func NewExternalBridgeService() ExternalBridgeService {
	return ExternalBridgeService{}
}

func (*ExternalBridgeService) readCompanyAndToken(ctx context.Context) (string, string) {
	var company = identitysdk.Empresa(ctx)
	var token = identitysdk.Token(ctx)
	return company, token
}
func (s *ExternalBridgeService) makeRequest(ctx context.Context, baseUrl, enpointPath, token string, v any) error {
	return s.makeRequestWithQueryPrams(ctx, baseUrl, enpointPath, token, nil, v)
}
func (*ExternalBridgeService) makeRequestWithQueryPrams(ctx context.Context, baseUrl, enpointPath, token string, queryParmas url.Values, v any) error {
	endpoint, err := url.JoinPath(baseUrl, enpointPath)
	if err != nil {
		slog.Error(
			"error joining base server url with `"+enpointPath+"`",
			"error", err,
			"baseurl", baseUrl,
		)
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		slog.Error(
			"error creating request",
			"error", err,
			"endpoint", endpoint,
		)
		return err
	}
	if queryParmas != nil {
		req.URL.RawQuery = queryParmas.Encode()
	}
	req.Header.Set("Authorization", token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("error on request", "error", err, "endpoint", endpoint, "method", "GET")
		return err
	}
	defer res.Body.Close()
	var jsondecoder = json.NewDecoder(res.Body)
	if res.StatusCode != http.StatusOK {
		var apiresponse struct {
			Message string `json:"message"`
		}
		if err := jsondecoder.Decode(&apiresponse); err != nil {
			slog.Error("error json decoding response", "error", err, "basepath", baseUrl, "path", enpointPath)
			return err
		}
		return errors.New(apiresponse.Message)
	}
	if err := jsondecoder.Decode(v); err != nil {
		slog.Error("error json decoding response", "error", err, "basepath", baseUrl, "path", enpointPath)
		return err
	}
	return err
}
