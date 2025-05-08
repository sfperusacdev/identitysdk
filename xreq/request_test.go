package xreq_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/sfperusacdev/identitysdk/xreq"
)

func TestExternalBridgeService_MakeRequest_AllOptions(t *testing.T) {
	expectedMethod := http.MethodPost
	expectedAuth := "Bearer testtoken"
	expectedHeaderKey := "X-Test-Header"
	expectedHeaderValue := "test-value"
	expectedQueryKey := "q"
	expectedQueryValue := "search"
	expectedBody := `{"input":"test"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate method
		if r.Method != expectedMethod {
			http.Error(w, `{"message":"invalid method"}`, http.StatusBadRequest)
			return
		}
		// Validate Authorization header
		if r.Header.Get("Authorization") != expectedAuth {
			http.Error(w, `{"message":"invalid authorization"}`, http.StatusBadRequest)
			return
		}
		// Validate custom header
		if r.Header.Get(expectedHeaderKey) != expectedHeaderValue {
			http.Error(w, `{"message":"missing header"}`, http.StatusBadRequest)
			return
		}
		// Validate query params
		if r.URL.Query().Get(expectedQueryKey) != expectedQueryValue {
			http.Error(w, `{"message":"missing query param"}`, http.StatusBadRequest)
			return
		}
		// Validate request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"message":"cannot read body"}`, http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		if string(body) != expectedBody {
			http.Error(w, `{"message":"body mismatch"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":"all_ok"}`))
	}))
	defer server.Close()

	var response struct {
		Data string `json:"data"`
	}

	query := url.Values{}
	query.Set(expectedQueryKey, expectedQueryValue)

	err := xreq.MakeRequest(
		context.Background(),
		server.URL,
		"/",
		xreq.WithMethod(expectedMethod),
		xreq.WithAuthorization(expectedAuth),
		xreq.WithHeader(expectedHeaderKey, expectedHeaderValue),
		xreq.WithQueryParams(query),
		xreq.WithRequestBody(strings.NewReader(expectedBody)),
		xreq.WithUnmarshalResponseInto(&response),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response.Data != "all_ok" {
		t.Fatalf("expected response data 'all_ok', got %q", response.Data)
	}
}

func TestExternalBridgeService_MakeRequest(t *testing.T) {
	type Response struct {
		Data string `json:"data"`
	}

	tests := []struct {
		name           string
		method         string
		path           string
		statusCode     int
		responseBody   any
		expectError    bool
		expectResponse *Response
	}{
		{
			name:           "GET success",
			method:         http.MethodGet,
			path:           "/get",
			statusCode:     http.StatusOK,
			responseBody:   Response{Data: "get_ok"},
			expectError:    false,
			expectResponse: &Response{Data: "get_ok"},
		},
		{
			name:           "POST success",
			method:         http.MethodPost,
			path:           "/post",
			statusCode:     http.StatusOK,
			responseBody:   Response{Data: "post_ok"},
			expectError:    false,
			expectResponse: &Response{Data: "post_ok"},
		},
		{
			name:           "PUT success",
			method:         http.MethodPut,
			path:           "/put",
			statusCode:     http.StatusOK,
			responseBody:   Response{Data: "put_ok"},
			expectError:    false,
			expectResponse: &Response{Data: "put_ok"},
		},
		{
			name:           "DELETE success",
			method:         http.MethodDelete,
			path:           "/delete",
			statusCode:     http.StatusOK,
			responseBody:   Response{Data: "delete_ok"},
			expectError:    false,
			expectResponse: &Response{Data: "delete_ok"},
		},
		{
			name:           "404 error",
			method:         http.MethodGet,
			path:           "/not-found",
			statusCode:     http.StatusNotFound,
			responseBody:   map[string]string{"message": "resource not found"},
			expectError:    true,
			expectResponse: nil,
		},
		{
			name:           "500 error",
			method:         http.MethodGet,
			path:           "/server-error",
			statusCode:     http.StatusInternalServerError,
			responseBody:   map[string]string{"message": "internal error"},
			expectError:    true,
			expectResponse: nil,
		},
		{
			name:           "invalid json response",
			method:         http.MethodGet,
			path:           "/bad-json",
			statusCode:     http.StatusOK,
			responseBody:   "invalid-json",
			expectError:    true,
			expectResponse: nil,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, tt := range tests {
			if r.Method == tt.method && r.URL.Path == tt.path {
				w.WriteHeader(tt.statusCode)
				switch body := tt.responseBody.(type) {
				case string:
					w.Write([]byte(body))
				default:
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(body)
				}
				return
			}
		}
		http.Error(w, `{"message":"unexpected error"}`, http.StatusInternalServerError)
	}))
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actualResponse Response

			err := xreq.MakeRequest(
				context.Background(),
				server.URL,
				tt.path,
				xreq.WithMethod(tt.method),
				xreq.WithUnmarshalResponseInto(&actualResponse),
			)

			if tt.expectError && err == nil {
				t.Fatalf("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Fatalf("expected no error but got %v", err)
			}

			if tt.expectResponse != nil && actualResponse != *tt.expectResponse {
				t.Fatalf("expected response %+v, got %+v", *tt.expectResponse, actualResponse)
			}
		})
	}
}

func TestExternalBridgeService_MakeRequest_WithCustomHeaders(t *testing.T) {
	expectedHeaderKey := "X-Custom-Header"
	expectedHeaderValue := "custom-value"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerValue := r.Header.Get(expectedHeaderKey)
		if headerValue != expectedHeaderValue {
			http.Error(w, `{"message":"missing or incorrect header"}`, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":"header_ok"}`))
	}))
	defer server.Close()

	var response struct {
		Data string `json:"data"`
	}

	err := xreq.MakeRequest(
		context.Background(),
		server.URL,
		"/",
		xreq.WithMethod(http.MethodGet),
		xreq.WithUnmarshalResponseInto(&response),
		xreq.WithHeader(expectedHeaderKey, expectedHeaderValue),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response.Data != "header_ok" {
		t.Fatalf("expected response data 'header_ok', got %q", response.Data)
	}
}

func TestExternalBridgeService_MakeRequest_WithRequestBody(t *testing.T) {
	expectedBody := `{"key":"value"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"message":"cannot read body"}`, http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		if string(body) != expectedBody {
			http.Error(w, `{"message":"body mismatch"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":"body_ok"}`))
	}))
	defer server.Close()

	var response struct {
		Data string `json:"data"`
	}

	err := xreq.MakeRequest(
		context.Background(),
		server.URL,
		"/",
		xreq.WithMethod(http.MethodPost),
		xreq.WithRequestBody(strings.NewReader(expectedBody)),
		xreq.WithUnmarshalResponseInto(&response),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response.Data != "body_ok" {
		t.Fatalf("expected response data 'body_ok', got %q", response.Data)
	}
}

func TestExternalBridgeService_MakeRequest_ErrorResponses(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		expectError  string
	}{
		{
			name:         "404 Not Found",
			statusCode:   http.StatusNotFound,
			responseBody: `{"message":"resource not found"}`,
			expectError:  "resource not found",
		},
		{
			name:         "500 Internal Server Error",
			statusCode:   http.StatusInternalServerError,
			responseBody: `{"message":"internal server error"}`,
			expectError:  "internal server error",
		},
		{
			name:         "400 Bad Request",
			statusCode:   http.StatusBadRequest,
			responseBody: `{"message":"bad request"}`,
			expectError:  "bad request",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			err := xreq.MakeRequest(
				context.Background(),
				server.URL,
				"/error",
				xreq.WithMethod(http.MethodGet),
			)

			if err == nil {
				t.Fatalf("expected an error but got nil")
			}

			if err.Error() != tt.expectError {
				t.Fatalf("expected error message %q, got %q", tt.expectError, err.Error())
			}
		})
	}
}
