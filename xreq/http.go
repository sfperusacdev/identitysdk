package xreq

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

type RequestOptions struct {
	Method       string
	QueryParams  url.Values
	Headers      http.Header
	RequestBody  io.Reader
	ResponseBody any // ResponseBody represents the full response body, used for all content types.
}

type RequestOption func(*RequestOptions)

func WithMethod(method string) RequestOption {
	return func(o *RequestOptions) {
		o.Method = method
	}
}

func WithAuthorization(token string) RequestOption {
	return WithHeader("Authorization", token)
}

func WithAccessToken(accessToken string) RequestOption {
	return WithHeader("X-Access-Token", accessToken)
}

func WithQueryParams(params url.Values) RequestOption {
	return func(o *RequestOptions) {
		o.QueryParams = params
	}
}

// WithUnmarshalResponseInto sets the full response body to be unmarshaled into the provided variable.
// Use this option when you need to capture the entire response, regardless of its structure.
func WithUnmarshalResponseInto(a any) RequestOption {
	return func(o *RequestOptions) {
		o.ResponseBody = a
	}
}

func WithJsonContentType() RequestOption {
	return WithHeader("Content-Type", "application/json")
}

func WithHeader(key, value string) RequestOption {
	return func(o *RequestOptions) {
		if o.Headers == nil {
			o.Headers = make(http.Header)
		}
		o.Headers.Set(key, value)

	}
}

func WithAddHeader(key, value string) RequestOption {
	return func(o *RequestOptions) {
		if o.Headers == nil {
			o.Headers = make(http.Header)
		}
		o.Headers.Add(key, value)
	}
}

func WithHeaders(headers http.Header) RequestOption {
	return func(o *RequestOptions) {
		o.Headers = headers
	}
}

func WithRequestBody(body io.Reader) RequestOption {
	return func(o *RequestOptions) {
		o.RequestBody = body
	}
}

func MakeRequest(ctx context.Context, baseUrl, endpointPath string, opts ...RequestOption) error {
	if baseUrl == "" {
		err := errors.New("base URL cannot be empty")
		slog.Error(
			"validating base URL",
			"error", err,
		)
		return err
	}
	options := &RequestOptions{
		Method: http.MethodGet,
	}
	for _, apply := range opts {
		apply(options)
	}
	endpoint, err := url.JoinPath(baseUrl, endpointPath)
	if err != nil {
		slog.Error(
			"joining base server url with `"+endpointPath+"`",
			"error", err,
			"baseurl", baseUrl,
		)
		return err
	}
	req, err := http.NewRequestWithContext(ctx, options.Method, endpoint, options.RequestBody)
	if err != nil {
		slog.Error(
			"error creating request",
			"error", err,
			"endpoint", endpoint,
		)
		return err
	}
	if options.QueryParams != nil {
		req.URL.RawQuery = options.QueryParams.Encode()
	}
	if options.Headers != nil {
		for key, values := range options.Headers {
			req.Header.Del(key)
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("error on request", "error", err, "endpoint", endpoint, "method", options.Method)
		return err
	}
	defer res.Body.Close()
	jsonDecoder := json.NewDecoder(res.Body)
	if res.StatusCode != http.StatusOK {
		var apiResponse struct {
			Message string `json:"message"`
		}
		if err := jsonDecoder.Decode(&apiResponse); err != nil {
			slog.Error("error json decoding response", "error", err, "basepath", baseUrl, "path", endpointPath)
			return err
		}
		slog.Error("service response error", "message", apiResponse.Message)
		return errors.New(apiResponse.Message)
	}
	if options.ResponseBody != nil {
		if err := jsonDecoder.Decode(options.ResponseBody); err != nil {
			slog.Error("error json decoding response", "error", err, "basepath", baseUrl, "path", endpointPath)
			return err
		}
	}
	return nil
}
