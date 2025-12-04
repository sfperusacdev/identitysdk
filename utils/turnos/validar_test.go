package turnos_test

import (
	"testing"

	"github.com/sfperusacdev/identitysdk/utils/turnos"
	"github.com/stretchr/testify/require"
)

func TestValidarTurnos_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		input       []turnos.Turno
		expectErr   bool
		expectCodes []string
	}{
		{
			name: "OK_DAY_NIGHT",
			input: []turnos.Turno{
				{"DAY", "06:00:00", "18:00:00"},
				{"NIGHT", "18:00:00", "06:00:00"},
			},
			expectErr:   false,
			expectCodes: []string{"DAY", "NIGHT"},
		},
		{
			name: "Overlap",
			input: []turnos.Turno{
				{"A", "06:00:00", "18:00:00"},
				{"B", "17:00:00", "06:00:00"},
			},
			expectErr: true,
		},
		{
			name: "Gap",
			input: []turnos.Turno{
				{"A", "06:00:00", "18:00:00"},
				{"B", "19:00:00", "06:00:00"},
			},
			expectErr: true,
		},
		{
			name: "Incomplete24h",
			input: []turnos.Turno{
				{"A", "06:00:00", "18:00:00"},
				{"B", "18:00:00", "05:00:00"},
			},
			expectErr: true,
		},
		{
			name: "InvalidTime",
			input: []turnos.Turno{
				{"A", "xx:00:00", "18:00:00"},
				{"B", "18:00:00", "06:00:00"},
			},
			expectErr: true,
		},
		{
			name: "SingleFullDay",
			input: []turnos.Turno{
				{"FULL", "00:00:00", "00:00:00"},
			},
			expectErr:   false,
			expectCodes: []string{"FULL"},
		},
		{
			name: "MultipleTurnsFullCycle",
			input: []turnos.Turno{
				{"T1", "00:00:00", "06:00:00"},
				{"T3", "12:00:00", "18:00:00"},
				{"T4", "18:00:00", "00:00:00"},
				{"T2", "06:00:00", "12:00:00"},
			},
			expectErr:   false,
			expectCodes: []string{"T1", "T2", "T3", "T4"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := turnos.ValidarTurnos(tc.input)

			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, len(tc.expectCodes), len(out))

			for i, code := range tc.expectCodes {
				require.Equal(t, code, out[i].Codigo)
			}
		})
	}
}

func TestValidarTurnos_Ordering(t *testing.T) {
	in := []turnos.Turno{
		{"T3", "12:00:00", "18:00:00"},
		{"T1", "00:00:00", "06:00:00"},
		{"T4", "18:00:00", "00:00:00"},
		{"T2", "06:00:00", "12:00:00"},
	}

	out, err := turnos.ValidarTurnos(in)
	require.NoError(t, err)

	require.Len(t, out, 4)

	require.Equal(t, "T1", out[0].Codigo)
	require.Equal(t, "T2", out[1].Codigo)
	require.Equal(t, "T3", out[2].Codigo)
	require.Equal(t, "T4", out[3].Codigo)
}
