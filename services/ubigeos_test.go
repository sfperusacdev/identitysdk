package services

import "testing"

func TestGetUbigeosMigration(t *testing.T) {
	const data = `[{"codigo":"150101","departamento_inei_nombre":"Lima","provincia_inei_nombre":"Lima","distrito_inei_nombre":"Lima"}]`
	checkMigration(t, migrationRequest{
		path: "/api/v2/ubigeos", resource: "com.sfperusac.general", data: data, want: data,
	}, func(f migrationFixture) ([]Ubigeo, error) {
		return f.legacy.GetUbigeos(f.ctx)
	}, func(f migrationFixture) ([]Ubigeo, error) {
		return f.global.GetUbigeos(f.ctx)
	})
}

func TestSearchUbigeoMigration(t *testing.T) {
	for _, source := range []UbigeoSource{UbigeoSourceINEI, UbigeoSourceRENIEC, UbigeoSourceSUNAT} {
		t.Run(string(source), func(t *testing.T) {
			for _, data := range []string{`{"codigo":"150101"}`, `null`} {
				t.Run(data, func(t *testing.T) {
					checkMigration(t, migrationRequest{
						path: "/api/v2/ubigeos/" + string(source) + "/Lima/Lima/Lima", resource: "com.sfperusac.general", data: data, want: data,
					}, func(f migrationFixture) (*Ubigeo, error) {
						return f.legacy.SearchUbigeo(f.ctx, source, "Lima", "Lima", "Lima")
					}, func(f migrationFixture) (*Ubigeo, error) {
						return f.global.SearchUbigeo(f.ctx, source, "Lima", "Lima", "Lima")
					})
				})
			}
		})
	}
}
