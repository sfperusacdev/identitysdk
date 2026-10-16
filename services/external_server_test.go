package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/configs"
	"github.com/sfperusacdev/identitysdk/sark_services/asistencia"
	"github.com/sfperusacdev/identitysdk/sark_services/contratos"
	"github.com/sfperusacdev/identitysdk/sark_services/global"
	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
)

type migrationFixture struct {
	ctx        context.Context
	legacy     *ExternalBridgeService
	contratos  *contratos.ContratosService
	global     *global.GlobalService
	identity   *bridgeidentity.IdentityService
	asistencia *asistencia.AsistenciaService
}

type migrationRequest struct {
	path         string
	resource     string
	query        map[string]string
	data         string
	want         string
	wantError    bool
	wantNotFound bool
}

func checkMigration[T any](t *testing.T, request migrationRequest, legacy, migrated func(migrationFixture) (T, error)) {
	t.Helper()
	for _, scenario := range []string{"response", "http error", "invalid json", "discovery error"} {
		if request.resource == "" && scenario == "discovery error" {
			continue
		}
		t.Run(scenario, func(t *testing.T) {
			var hits int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/api/v1/get-service-location" {
					if r.URL.Query().Get("company_code") != "empresa-test" || r.URL.Query().Get("resource_code") != request.resource {
						t.Errorf("unexpected discovery query: %s", r.URL.RawQuery)
					}
					if scenario == "discovery error" {
						w.WriteHeader(http.StatusInternalServerError)
					}
					fmt.Fprintf(w, `{"data":{"location":"http://%s"}}`, r.Host)
					return
				}
				hits++
				if r.Method != http.MethodGet || r.URL.Path != request.path {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
				}
				for key, value := range request.query {
					if got := r.URL.Query().Get(key); got != value {
						t.Errorf("query %s = %q, want %q", key, got, value)
					}
				}
				if request.resource == "" {
					if r.Header.Get("X-Access-Token") != "access-token" {
						t.Error("missing identity access token")
					}
				} else if r.Header.Get("Authorization") != "Bearer session-token" {
					t.Error("missing session token")
				}
				switch scenario {
				case "http error":
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprint(w, `{"message":"not found"}`)
				case "invalid json":
					fmt.Fprint(w, `{"data":`)
				default:
					fmt.Fprintf(w, `{"data":%s}`, request.data)
				}
			}))
			t.Cleanup(server.Close)
			previousServer := identitysdk.GetIdentityServer()
			identitysdk.SetIdentityServer(server.URL)
			t.Cleanup(func() { identitysdk.SetIdentityServer(previousServer) })
			conf := &configs.GeneralServiceConfig{IdentityValue: server.URL, IdentityAccessTokenValue: "access-token"}
			env := bridgeidentity.NewIdentityProvider()
			ctx := identitysdk.CtxWithToken(identitysdk.CtxWithDomain(t.Context(), "empresa-test"), "Bearer session-token")
			fixture := migrationFixture{
				ctx: ctx, legacy: NewExternalBridgeService(conf),
				contratos: contratos.NewContratosService(env), global: global.NewGlobalService(env),
				identity: bridgeidentity.NewIdentityService(conf, env), asistencia: asistencia.NewAsistenciaService(nil),
			}
			for _, implementation := range []struct {
				name string
				call func(migrationFixture) (T, error)
			}{{"legacy", legacy}, {"migrated", migrated}} {
				t.Run(implementation.name, func(t *testing.T) {
					got, err := implementation.call(fixture)
					wantError := scenario != "response" || request.wantError
					if (err != nil) != wantError {
						t.Fatalf("error = %v, wantError %v", err, wantError)
					}
					if wantError {
						var zero T
						if !reflect.DeepEqual(got, zero) {
							t.Errorf("result on error = %+v, want zero value", got)
						}
						if scenario == "response" && request.wantNotFound && !errors.Is(err, contratos.ErrNotFound) {
							t.Errorf("error = %v, want contratos.ErrNotFound", err)
						}
						return
					}
					var want T
					if err := json.Unmarshal([]byte(request.want), &want); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(got, want) {
						t.Errorf("result = %+v, want %+v", got, want)
					}
				})
			}
			wantHits := 2
			if scenario == "discovery error" {
				wantHits = 0
			}
			if hits != wantHits {
				t.Errorf("API requests = %d, want %d", hits, wantHits)
			}
		})
	}
}

func TestErrNotFoundCompatibility(t *testing.T) {
	if ErrNotFound != contratos.ErrNotFound {
		t.Fatal("legacy and migrated ErrNotFound must be identical")
	}
}
