package global

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sfperusacdev/identitysdk"
	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
)

func TestGlobalService_GetWorkflow(t *testing.T) {
	tests := []struct {
		name            string
		locationStatus  int
		workflowStatus  int
		body            string
		wantError       bool
		wantWorkflowHit bool
	}{
		{
			name:            "success",
			locationStatus:  http.StatusOK,
			workflowStatus:  http.StatusOK,
			body:            `{"message":"ok","data":{"codigo":"factura","descripcion":"Factura","workflow_state":[{"codigo":"pending","nombre":"Pendiente"}],"workflow_transition":[{"codigo":"approve","origen":"pending","destino":"approved"}]}}`,
			wantWorkflowHit: true,
		},
		{
			name:           "service discovery error",
			locationStatus: http.StatusInternalServerError,
			wantError:      true,
		},
		{
			name:            "workflow request error",
			locationStatus:  http.StatusOK,
			workflowStatus:  http.StatusNotFound,
			body:            `{"message":"workflow not found"}`,
			wantError:       true,
			wantWorkflowHit: true,
		},
		{
			name:            "invalid workflow response",
			locationStatus:  http.StatusOK,
			workflowStatus:  http.StatusOK,
			body:            `{"data":`,
			wantError:       true,
			wantWorkflowHit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflowHit := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/api/v1/get-service-location":
					if got := r.URL.Query().Get("company_code"); got != "empresa-test" {
						t.Errorf("company_code = %q, want empresa-test", got)
					}
					if got := r.URL.Query().Get("resource_code"); got != "com.sfperusac.general" {
						t.Errorf("resource_code = %q, want com.sfperusac.general", got)
					}
					w.WriteHeader(tt.locationStatus)
					fmt.Fprintf(w, `{"data":{"location":"http://%s"}}`, r.Host)
				case "/api/v2/documentos/workflows/factura":
					workflowHit = true
					if r.Method != http.MethodGet {
						t.Errorf("method = %q, want GET", r.Method)
					}
					if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
						t.Errorf("Authorization = %q, want Bearer test-token", got)
					}
					if got := r.Header.Get("Content-Type"); got != "application/json" {
						t.Errorf("Content-Type = %q, want application/json", got)
					}
					w.WriteHeader(tt.workflowStatus)
					fmt.Fprint(w, tt.body)
				default:
					t.Errorf("unexpected request: %s", r.URL)
					http.NotFound(w, r)
				}
			}))
			t.Cleanup(server.Close)

			previousServer := identitysdk.GetIdentityServer()
			identitysdk.SetIdentityServer(server.URL)
			t.Cleanup(func() { identitysdk.SetIdentityServer(previousServer) })

			ctx := identitysdk.CtxWithDomain(t.Context(), "empresa-test")
			ctx = identitysdk.CtxWithToken(ctx, "Bearer test-token")
			service := NewGlobalService(bridgeidentity.NewIdentityProvider())
			workflow, err := service.GetWorkflow(ctx, "factura")
			if (err != nil) != tt.wantError {
				t.Fatalf("GetWorkflow() error = %v, wantError %v", err, tt.wantError)
			}
			if workflowHit != tt.wantWorkflowHit {
				t.Errorf("workflow requested = %v, want %v", workflowHit, tt.wantWorkflowHit)
			}
			if tt.wantError {
				if workflow != nil {
					t.Errorf("GetWorkflow() = %+v, want nil on error", workflow)
				}
				return
			}
			if workflow == nil {
				t.Fatal("GetWorkflow() returned nil workflow")
			}
			if workflow.Codigo != "factura" || workflow.Descripcion != "Factura" {
				t.Errorf("unexpected workflow: %+v", workflow)
			}
			if len(workflow.WorkflowState) != 1 || workflow.WorkflowState[0].Codigo != "pending" {
				t.Errorf("unexpected workflow states: %+v", workflow.WorkflowState)
			}
			if len(workflow.WorkflowTransition) != 1 || workflow.WorkflowTransition[0].Destino != "approved" {
				t.Errorf("unexpected workflow transitions: %+v", workflow.WorkflowTransition)
			}
		})
	}
}
