package services

import "testing"

func TestGetPosicionesActivasMigration(t *testing.T) {
	checkMigration(t, migrationRequest{
		path: "/v1/fotocheck/resumen/puestos", resource: "com.sfperusac.contratos",
		data: `["cargo-1","cargo-2"]`, want: `["cargo-1","cargo-2"]`,
	}, func(f migrationFixture) ([]string, error) {
		return f.legacy.GetPosicionesActivas(f.ctx)
	}, func(f migrationFixture) ([]string, error) {
		return f.contratos.GetPosicionesActivas(f.ctx)
	})
}
