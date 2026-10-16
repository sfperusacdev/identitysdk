package services

import "testing"

func TestGetTrabajadorPosicionMigration(t *testing.T) {
	for _, tt := range []struct {
		name      string
		data      string
		wantError bool
	}{
		{name: "found", data: `[{"cargo_codigo":"cargo-1"}]`},
		{name: "not found", data: `[]`, wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			checkMigration(t, migrationRequest{
				path: "/v1/fotocheck/trabajador-1/json", resource: "com.sfperusac.contratos",
				data: tt.data, want: `"cargo-1"`, wantError: tt.wantError, wantNotFound: tt.wantError,
			}, func(f migrationFixture) (string, error) {
				return f.legacy.GetTrabajadorPosicion(f.ctx, "trabajador-1")
			}, func(f migrationFixture) (string, error) {
				return f.contratos.GetTrabajadorPosicion(f.ctx, "trabajador-1")
			})
		})
	}
}
