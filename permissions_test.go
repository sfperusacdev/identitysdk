package identitysdk_test

import (
	"context"
	"testing"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/entities"
)

func TestHasPermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		permissions []entities.Permission
		branch      string
		noSession   bool
		want        bool
	}{
		{name: "no session", noSession: true, branch: "01"},
		{name: "no permissions", branch: "01"},
		{name: "no branches", permissions: []entities.Permission{{ID: "read"}}, branch: "01"},
		{name: "explicit branch", permissions: []entities.Permission{{ID: "read", CompanyBrances: []string{"01"}}}, branch: "01", want: true},
		{name: "different branch", permissions: []entities.Permission{{ID: "read", CompanyBrances: []string{"01"}}}, branch: "02"},
		{name: "all branches", permissions: []entities.Permission{{ID: "read", CompanyBrances: []string{"*"}}}, branch: "02", want: true},
		{name: "wildcard alongside explicit branch", permissions: []entities.Permission{{ID: "read", CompanyBrances: []string{"01", "*"}}}, branch: "02", want: true},
		{name: "wildcard for another permission", permissions: []entities.Permission{{ID: "write", CompanyBrances: []string{"*"}}}, branch: "02"},
		{name: "wildcard does not expand permission IDs", permissions: []entities.Permission{{ID: "*", CompanyBrances: []string{"*"}}}, branch: "01"},
		{name: "later matching grant", permissions: []entities.Permission{{ID: "read", CompanyBrances: []string{"01"}}, {ID: "read", CompanyBrances: []string{"*"}}}, branch: "02", want: true},
		{name: "admin is not an implicit bypass", permissions: []entities.Permission{{ID: "admin", CompanyBrances: []string{"*"}}}, branch: "01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			if !tt.noSession {
				ctx = identitysdk.CtxWithSession(ctx, entities.Session{Permissions: tt.permissions})
			}
			if got := identitysdk.HasPermission(ctx, "read", tt.branch); got != tt.want {
				t.Errorf("HasPermission(ctx, read, %q) = %v, want %v", tt.branch, got, tt.want)
			}
		})
	}
}
