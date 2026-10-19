package permissions_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/entities"
	"github.com/sfperusacdev/identitysdk/permissions"
)

func TestNewPermissionMiddlewareBuilder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		permissions []entities.Permission
		wantAllowed bool
	}{
		{name: "no permissions"},
		{name: "explicit branch", permissions: []entities.Permission{{ID: "read", CompanyBrances: []string{"02"}}}, wantAllowed: true},
		{name: "different branch", permissions: []entities.Permission{{ID: "read", CompanyBrances: []string{"01"}}}},
		{name: "all branches", permissions: []entities.Permission{{ID: "read", CompanyBrances: []string{"*"}}}, wantAllowed: true},
		{name: "wildcard for another permission", permissions: []entities.Permission{{ID: "write", CompanyBrances: []string{"*"}}}},
		{name: "admin all branches", permissions: []entities.Permission{{ID: "admin", CompanyBrances: []string{"*"}}}, wantAllowed: true},
		{name: "admin different branch", permissions: []entities.Permission{{ID: "admin", CompanyBrances: []string{"01"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := identitysdk.CtxWithSucursal(req.Context(), "02")
			ctx = identitysdk.CtxWithSession(ctx, entities.Session{Permissions: tt.permissions})
			recorder := httptest.NewRecorder()
			c := e.NewContext(req.WithContext(ctx), recorder)
			called := false
			handler := permissions.NewPermissionMiddlewareBuilder()([]string{"read"})(func(c echo.Context) error {
				called = true
				return c.NoContent(http.StatusNoContent)
			})
			if err := handler(c); err != nil {
				e.HTTPErrorHandler(err, c)
			}
			if called != tt.wantAllowed {
				t.Errorf("handler called = %v, want %v", called, tt.wantAllowed)
			}
			wantStatus := http.StatusForbidden
			if tt.wantAllowed {
				wantStatus = http.StatusNoContent
			}
			if recorder.Code != wantStatus {
				t.Errorf("status = %d, want %d; body: %s", recorder.Code, wantStatus, recorder.Body.String())
			}
		})
	}
}
