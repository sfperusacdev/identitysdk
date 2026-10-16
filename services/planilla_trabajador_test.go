package services

import (
	"testing"

	"github.com/sfperusacdev/identitysdk/entities"
)

func TestGetPlanillasMigration(t *testing.T) {
	const data = `[{"codigo":"planilla-1","descripcion":"Planilla","default_grupo_provision_codigo":"grupo-1"}]`
	checkMigration(t, migrationRequest{
		path: "/v1/api/tables/empresa/planilla", resource: "com.sfperusac.contratos", data: data, want: data,
	}, func(f migrationFixture) ([]PlanillaDto, error) {
		return f.legacy.GetPlanillas(f.ctx)
	}, func(f migrationFixture) ([]PlanillaDto, error) {
		return f.contratos.GetPlanillas(f.ctx)
	})
}

func TestGetPlanillaTrabajadorMigration(t *testing.T) {
	const data = `{"codigo":"trabajador-1","planilla_codigo":"planilla-1","grupo_provicion_codigo":"grupo-1"}`
	checkMigration(t, migrationRequest{
		path: "/v1/trabajadores/planilla/empresa-test/trabajador-1", resource: "com.sfperusac.contratos", data: data, want: data,
	}, func(f migrationFixture) (*entities.TrabajadorPlanilla, error) {
		return f.legacy.GetPlanillaTrabajador(f.ctx, "empresa-test.trabajador-1")
	}, func(f migrationFixture) (*entities.TrabajadorPlanilla, error) {
		return f.contratos.GetPlanillaTrabajador(f.ctx, "empresa-test.trabajador-1")
	})
}
