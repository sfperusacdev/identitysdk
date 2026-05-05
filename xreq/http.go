package xreq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/user0608/goones/errs"
)

type XReqBody struct {
	Error  error
	Reader io.Reader
}
type RequestOptions struct {
	Method       string
	QueryParams  url.Values
	Headers      http.Header
	RequestBody  XReqBody
	ResponseBody any // ResponseBody is decoded from a JSON response body.
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

func WithQueryParam(key string, value string) RequestOption {
	return func(o *RequestOptions) {
		if o.QueryParams == nil {
			o.QueryParams = make(url.Values)
		}
		o.QueryParams.Set(key, value)
	}
}

func WithQueryParams(params url.Values) RequestOption {
	return func(o *RequestOptions) {
		if o.QueryParams == nil {
			o.QueryParams = make(url.Values)
		}
		for k, v := range params {
			for _, val := range v {
				o.QueryParams.Add(k, val)
			}
		}
	}
}

func WithJSONBody(payload any) RequestOption {
	return func(ro *RequestOptions) {
		body, err := json.Marshal(payload)
		if err != nil {
			ro.RequestBody = XReqBody{
				Error: err,
			}
			return
		}
		ro.RequestBody = XReqBody{Reader: bytes.NewReader(body)}
		if ro.Headers == nil {
			ro.Headers = make(http.Header)
		}
		ro.Headers.Set("Content-Type", "application/json")
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

func WithXOrigin(value string) RequestOption {
	return WithHeader("X-Origin", value)
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
		if o.Headers == nil {
			o.Headers = make(http.Header)
		}

		for key, values := range headers {
			o.Headers.Del(key)
			for _, value := range values {
				o.Headers.Add(key, value)
			}
		}
	}
}

func WithRequestBody(body io.Reader) RequestOption {
	return func(o *RequestOptions) {
		o.RequestBody = XReqBody{Reader: body}
	}
}

var defaultXOrigin string

func SetDefaultXOrigin(xorigin string) { defaultXOrigin = xorigin }

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
		return errs.InternalError(err, "falló la construcción de la URL para %s%s", baseUrl, endpointPath)
	}
	if options.RequestBody.Error != nil {
		return errs.InternalError(
			options.RequestBody.Error,
			"no se pudo preparar el body de la solicitud para %s %s",
			options.Method,
			endpoint,
		)
	}
	req, err := http.NewRequestWithContext(ctx, options.Method, endpoint, options.RequestBody.Reader)
	if err != nil {
		slog.Error(
			"error creating request",
			"error", err,
			"endpoint", endpoint,
		)
		return errs.InternalError(err, "falló la creación de la request para %s", endpoint)
	}

	if options.QueryParams != nil {
		req.URL.RawQuery = options.QueryParams.Encode()
	}

	if defaultXOrigin != "" {
		if options.Headers == nil {
			options.Headers = make(http.Header)
		}
		options.Headers.Set("X-Origin", defaultXOrigin)
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
		return errs.InternalError(err, "falló la petición %s a %s", options.Method, endpoint)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			return errs.InternalError(readErr, "falló la lectura de la respuesta de error de %s", endpoint)
		}

		var apiResponse struct {
			Message string `json:"message"`
		}

		message := string(body)
		if err := json.Unmarshal(body, &apiResponse); err == nil && apiResponse.Message != "" {
			message = apiResponse.Message
		}

		slog.Error(
			"service response error",
			"status", res.StatusCode,
			"message", message,
			"endpoint", endpoint,
			"method", options.Method,
		)

		return errs.InternalErrorDirect(message)
	}

	if options.ResponseBody != nil {
		jsonDecoder := json.NewDecoder(res.Body)
		if err := jsonDecoder.Decode(options.ResponseBody); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			slog.Error("error json decoding response", "error", err, "basepath", baseUrl, "path", endpointPath)
			return errs.InternalError(
				err,
				"falló la decodificación de la respuesta JSON de %s%s",
				baseUrl,
				endpointPath,
			)
		}
	}
	return nil
}
