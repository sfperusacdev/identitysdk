package contratos

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestResumenTrabajadorFotocheckDTO_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		fecha   string
		want    string
		wantErr bool
	}{
		{name: "vacía", fecha: `,"fecha_ingreso":""`},
		{name: "null", fecha: `,"fecha_ingreso":null`},
		{name: "ausente"},
		{name: "solo fecha", fecha: `,"fecha_ingreso":"2026-09-01"`, want: "2026-09-01T00:00:00Z"},
		{name: "RFC3339", fecha: `,"fecha_ingreso":"2026-09-01T08:30:00-05:00"`, want: "2026-09-01T08:30:00-05:00"},
		{name: "fracciones", fecha: `,"fecha_ingreso":"2026-09-01T08:30:00.123Z"`, want: "2026-09-01T08:30:00.123Z"},
		{name: "fecha inválida", fecha: `,"fecha_ingreso":"2026-02-30"`, wantErr: true},
		{name: "texto inválido", fecha: `,"fecha_ingreso":"no es fecha"`, wantErr: true},
		{name: "tipo inválido", fecha: `,"fecha_ingreso":123`, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var dto ResumenTrabajadorFotocheckDTO
			err := json.Unmarshal([]byte(`{"codigo":"empresa.w1","planilla_codigo":"empresa.P1","estado":true,"fecha_inicio_ultimo_contrato":"2026-09-01T00:00:00Z"`+tc.fecha+`}`), &dto)
			if tc.wantErr {
				if err == nil || !strings.Contains(err.Error(), "fecha_ingreso") {
					t.Fatalf("esperaba error de fecha_ingreso, recibió %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if dto.Codigo != "empresa.w1" || dto.PlanillaCodigo != "empresa.P1" || !dto.Estado || dto.FechaInicioUltimoContrato == nil {
				t.Fatalf("se perdieron campos del trabajador: %+v", dto)
			}
			if tc.want == "" {
				if dto.FechaIngreso != nil {
					t.Fatalf("esperaba nil, recibió %v", dto.FechaIngreso)
				}
			} else if dto.FechaIngreso == nil || dto.FechaIngreso.Format(time.RFC3339Nano) != tc.want {
				t.Fatalf("fecha incorrecta: %v; esperaba %s", dto.FechaIngreso, tc.want)
			}
		})
	}
}

func TestResumenTrabajadorFotocheckDTO_JSONRoundTrip(t *testing.T) {
	t.Parallel()
	fecha := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	original := ResumenTrabajadorFotocheckDTO{Codigo: "empresa.w1", FechaIngreso: &fecha}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var dto ResumenTrabajadorFotocheckDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		t.Fatal(err)
	}
	if dto.Codigo != original.Codigo || dto.FechaIngreso == nil || !dto.FechaIngreso.Equal(fecha) {
		t.Fatalf("round trip incorrecto: %+v", dto)
	}
	if err := json.Unmarshal([]byte(`{"fecha_ingreso":""}`), &dto); err != nil {
		t.Fatal(err)
	}
	if dto.FechaIngreso != nil {
		t.Fatal("una fecha vacía debe restablecer FechaIngreso a nil")
	}
}
