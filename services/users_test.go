package services

import (
	"testing"

	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
)

func TestGetUsuariosMigration(t *testing.T) {
	const data = `[{"code":"user-1","username":"ana","subordinates":[{"code":"user-2","username":"luis"}]}]`
	checkMigration(t, migrationRequest{
		path: "/_/api/v1/internal/system/users", query: map[string]string{"company_code": "empresa-test"}, data: data, want: data,
	}, func(f migrationFixture) ([]UserAccountWithSubordinates, error) {
		return f.legacy.GetUsuarios(f.ctx, "  empresa-test  ")
	}, func(f migrationFixture) ([]UserAccountWithSubordinates, error) {
		return f.identity.GetUsuarios(f.ctx, "  empresa-test  ")
	})
	t.Run("empty domain without configuration", func(t *testing.T) {
		service := bridgeidentity.NewIdentityService(nil, bridgeidentity.NewIdentityProvider())
		got, err := service.GetUsuarios(t.Context(), "  ")
		if err != nil || got == nil || len(got) != 0 {
			t.Fatalf("GetUsuarios(empty) = %+v, %v; want empty non-nil slice", got, err)
		}
	})
}

func TestGetGruposMigration(t *testing.T) {
	const data = `[{"code":"grupo-1","description":"Administradores","users":[{"code":"user-1","username":"ana"}]}]`
	checkMigration(t, migrationRequest{
		path: "/_/api/v1/internal/system/groups", query: map[string]string{"company_code": "empresa-test"}, data: data, want: data,
	}, func(f migrationFixture) ([]GrupoWithUsers, error) {
		return f.legacy.GetGrupos(f.ctx, "  empresa-test  ")
	}, func(f migrationFixture) ([]GrupoWithUsers, error) {
		return f.identity.GetGrupos(f.ctx, "  empresa-test  ")
	})
	t.Run("empty domain without configuration", func(t *testing.T) {
		service := bridgeidentity.NewIdentityService(nil, bridgeidentity.NewIdentityProvider())
		got, err := service.GetGrupos(t.Context(), "  ")
		if err != nil || got == nil || len(got) != 0 {
			t.Fatalf("GetGrupos(empty) = %+v, %v; want empty non-nil slice", got, err)
		}
	})
}
