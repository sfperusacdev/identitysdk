package templates

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sfperusacdev/identitysdk"
	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
)

var testID = uuid.MustParse("67d06a72-98ad-44e8-962c-609401a3f46d")

func testClient(t *testing.T, handler http.HandlerFunc) *TemplatesService {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/get-service-location" {
			if r.URL.Query().Get("company_code") != "acme" ||
				r.URL.Query().Get("resource_code") != "com.sfperusac.general" {
				t.Errorf("discovery incorrecto: %s", r.URL)
			}
			fmt.Fprintf(w, `{"data":{"location":"%s"}}`, serverURL(r))
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	previous := identitysdk.GetIdentityServer()
	identitysdk.SetIdentityServer(server.URL)
	t.Cleanup(func() { identitysdk.SetIdentityServer(previous) })
	return NewTemplatesService(bridgeidentity.NewIdentityProvider())
}

func serverURL(r *http.Request) string { return "http://" + r.Host }

func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx := identitysdk.CtxWithDomain(t.Context(), "acme")
	return identitysdk.CtxWithToken(ctx, "Bearer test-token")
}

func TestTemplatesService_Get(t *testing.T) {
	for _, tt := range []struct {
		name   string
		status int
		want   string
	}{
		{name: "archivo", status: http.StatusOK, want: "contenido docx"},
		{name: "no encontrado", status: http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/v1/api/templates/"+testID.String() {
					t.Errorf("petición inesperada: %s %s", r.Method, r.URL)
				}
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.want)
			})
			got, err := client.Get(testContext(t), testID)
			if (err != nil) != (tt.status != http.StatusOK) {
				t.Fatalf("Get() error = %v", err)
			}
			if err == nil && string(got) != tt.want {
				t.Errorf("Get() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTemplatesService_Put(t *testing.T) {
	for _, tt := range []struct {
		name   string
		status int
	}{
		{name: "sube archivo", status: http.StatusOK},
		{name: "error del servidor", status: http.StatusBadRequest},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut || r.URL.Path != "/v1/api/templates/"+testID.String() {
					t.Errorf("petición inesperada: %s %s", r.Method, r.URL)
				}
				file, header, err := r.FormFile("file")
				if err != nil {
					t.Errorf("FormFile(file): %v", err)
				} else {
					defer file.Close()
					content, err := io.ReadAll(file)
					if err != nil || !bytes.Equal(content, []byte("docx")) {
						t.Errorf("archivo = %q, error = %v", content, err)
					}
					if header.Filename != testID.String()+".docx" {
						t.Errorf("nombre de archivo = %q", header.Filename)
					}
				}
				w.WriteHeader(tt.status)
			})
			err := client.Put(testContext(t), testID, []byte("docx"))
			if (err != nil) != (tt.status != http.StatusOK) {
				t.Fatalf("Put() error = %v", err)
			}
		})
	}
}

func TestTemplatesService_Delete(t *testing.T) {
	for _, tt := range []struct {
		name   string
		status int
	}{
		{name: "elimina", status: http.StatusOK},
		{name: "no encontrado", status: http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete || r.URL.Path != "/v1/api/templates/"+testID.String() {
					t.Errorf("petición inesperada: %s %s", r.Method, r.URL)
				}
				w.WriteHeader(tt.status)
			})
			err := client.Delete(testContext(t), testID)
			if (err != nil) != (tt.status != http.StatusOK) {
				t.Fatalf("Delete() error = %v", err)
			}
		})
	}
}

func TestTemplatesService_GetHash(t *testing.T) {
	for _, tt := range []struct {
		name      string
		status    int
		body      string
		want      string
		wantError bool
	}{
		{name: "hash", status: http.StatusOK, body: `{"data":"sha256"}`, want: "sha256"},
		{name: "no encontrado", status: http.StatusNotFound, wantError: true},
		{name: "respuesta inválida", status: http.StatusOK, body: `{"data":`, wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/v1/api/templates/"+testID.String()+"/hash" {
					t.Errorf("petición inesperada: %s %s", r.Method, r.URL)
				}
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			})
			got, err := client.GetHash(testContext(t), testID)
			if (err != nil) != tt.wantError {
				t.Fatalf("GetHash() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("GetHash() = %q, want %q", got, tt.want)
			}
		})
	}
}
