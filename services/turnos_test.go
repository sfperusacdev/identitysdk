package services

import (
	"testing"

	"github.com/sfperusacdev/identitysdk/utils/turnos"
)

func TestTurnosMigration(t *testing.T) {
	const day = `{"codigo":"dia","inicio":"06:00:00","fin":"18:00:00","referecia_externa":"D"}`
	const night = `{"codigo":"noche","inicio":"18:00:00","fin":"06:00:00","referecia_externa":"N"}`
	for _, tt := range []struct {
		name      string
		data      string
		wantError bool
	}{
		{name: "sorted full day", data: "[" + night + "," + day + "]"},
		{name: "empty", data: `[]`, wantError: true},
		{name: "incomplete coverage", data: "[" + day + "]", wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			checkMigration(t, migrationRequest{
				path: "/v1/api/turnos", resource: "com.sfperusac.asistencia", query: map[string]string{"utc": "1"},
				data: tt.data, want: "[" + day + "," + night + "]", wantError: tt.wantError,
			}, func(f migrationFixture) ([]turnos.Turno, error) {
				return f.legacy.Turnos(f.ctx)
			}, func(f migrationFixture) ([]turnos.Turno, error) {
				return f.asistencia.Turnos(f.ctx)
			})
		})
	}
}
