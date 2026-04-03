package sqlproc

import (
	"errors"
	"testing"
)

func TestGetStoredProcedureIdentifierFromQuery(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expected    string
		expectedErr error
	}{
		{
			name:     "exec simple one line",
			query:    "EXEC dbo.ObtenerUsuarios @Id = 1",
			expected: "dbo.ObtenerUsuarios",
		},
		{
			name:     "execute simple one line",
			query:    "EXECUTE dbo.ObtenerUsuarios @Id = 1",
			expected: "dbo.ObtenerUsuarios",
		},
		{
			name:     "exec without schema",
			query:    "EXEC ObtenerUsuarios @Id = 1",
			expected: "ObtenerUsuarios",
		},
		{
			name:     "exec with bracketed identifier",
			query:    "EXEC [dbo].[ObtenerUsuarios] @Id = 1",
			expected: "[dbo].[ObtenerUsuarios]",
		},
		{
			name:     "exec with fully qualified identifier",
			query:    "EXEC servidor.base.dbo.ObtenerUsuarios @Id = 1",
			expected: "servidor.base.dbo.ObtenerUsuarios",
		},
		{
			name:     "exec with extra spaces",
			query:    "   EXEC    dbo.ObtenerUsuarios    @Id = 1   ",
			expected: "dbo.ObtenerUsuarios",
		},
		{
			name:     "exec in multiline",
			query:    "EXEC\n    dbo.ObtenerUsuarios\n    @Id = 1,\n    @Activo = 1",
			expected: "dbo.ObtenerUsuarios",
		},
		{
			name:     "execute in multiline with bracketed identifier",
			query:    "EXECUTE\n    [dbo].[ObtenerUsuarios]\n    @Id = 1",
			expected: "[dbo].[ObtenerUsuarios]",
		},
		{
			name:     "exec with line comment before identifier",
			query:    "EXEC -- comentario\n dbo.ObtenerUsuarios @Id = 1",
			expected: "dbo.ObtenerUsuarios",
		},
		{
			name:     "exec with block comment before identifier",
			query:    "EXEC /* comentario */ dbo.ObtenerUsuarios @Id = 1",
			expected: "dbo.ObtenerUsuarios",
		},
		{
			name:     "exec with comments between identifier parts",
			query:    "EXEC [dbo] /* x */ . /* y */ [ObtenerUsuarios] @Id = 1",
			expected: "[dbo].[ObtenerUsuarios]",
		},
		{
			name:     "exec with multiline block comments between parts",
			query:    "EXEC\n[dbo]\n/* bloque\ncomentario */\n.\n/* otro */\n[ObtenerUsuarios]\n@Id = 1",
			expected: "[dbo].[ObtenerUsuarios]",
		},
		{
			name:     "exec with line comments between identifier parts",
			query:    "EXEC [dbo] -- schema\n . -- dot\n [ObtenerUsuarios] @Id = 1",
			expected: "[dbo].[ObtenerUsuarios]",
		},
		{
			name:     "exec lowercase",
			query:    "exec dbo.ObtenerUsuarios @Id = 1",
			expected: "dbo.ObtenerUsuarios",
		},
		{
			name:     "execute mixed case",
			query:    "ExEcUtE dbo.ObtenerUsuarios @Id = 1",
			expected: "dbo.ObtenerUsuarios",
		},
		{
			name:     "exec with tabs and newlines",
			query:    "\tEXEC\t[dbo].\t[ObtenerUsuarios]\n\t@Id = 1",
			expected: "[dbo].[ObtenerUsuarios]",
		},
		{
			name:        "comment only does not contain executable procedure",
			query:       "-- EXEC dbo.ObtenerUsuarios",
			expectedErr: ErrStoredProcedureNotFound,
		},
		{
			name:        "block comment only does not contain executable procedure",
			query:       "/* EXEC dbo.ObtenerUsuarios */",
			expectedErr: ErrStoredProcedureNotFound,
		},
		{
			name:     "fake commented exec then real exec",
			query:    "-- EXEC dbo.Fake\nEXEC dbo.Real @Id = 1",
			expected: "dbo.Real",
		},
		{
			name:        "invalid query select",
			query:       "SELECT * FROM Usuarios",
			expectedErr: ErrStoredProcedureNotFound,
		},
		{
			name:        "invalid query empty",
			query:       "",
			expectedErr: ErrStoredProcedureNotFound,
		},
		{
			name:        "invalid query only exec",
			query:       "EXEC",
			expectedErr: ErrStoredProcedureNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := GetStoredProcedureIdentifierFromQuery(tt.query)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if actual != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

func TestReplaceStoredProcedureIdentifierInQuery(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		newIdentifier string
		expected      string
		expectedErr   error
	}{
		{
			name:          "replace simple one line",
			query:         "EXEC dbo.ObtenerUsuarios @Id = 1",
			newIdentifier: "[audit].[ListarUsuarios]",
			expected:      "EXEC [audit].[ListarUsuarios] @Id = 1",
		},
		{
			name:          "replace without schema",
			query:         "EXEC ObtenerUsuarios @Id = 1",
			newIdentifier: "nuevo_esquema.ListarUsuarios",
			expected:      "EXEC nuevo_esquema.ListarUsuarios @Id = 1",
		},
		{
			name:          "replace bracketed identifier",
			query:         "EXEC [dbo].[ObtenerUsuarios] @Id = 1",
			newIdentifier: "[nuevo].[ListarUsuarios]",
			expected:      "EXEC [nuevo].[ListarUsuarios] @Id = 1",
		},
		{
			name:          "replace fully qualified identifier",
			query:         "EXEC servidor.base.dbo.ObtenerUsuarios @Id = 1",
			newIdentifier: "[otro].[NuevoSP]",
			expected:      "EXEC [otro].[NuevoSP] @Id = 1",
		},
		{
			name:          "replace in multiline",
			query:         "EXEC\n    dbo.ObtenerUsuarios\n    @Id = 1,\n    @Activo = 1",
			newIdentifier: "dbo.ListarUsuarios",
			expected:      "EXEC\n    dbo.ListarUsuarios\n    @Id = 1,\n    @Activo = 1",
		},
		{
			name:          "replace with line comment before identifier",
			query:         "EXEC -- comentario\n dbo.ObtenerUsuarios @Id = 1",
			newIdentifier: "[audit].[ListarUsuarios]",
			expected:      "EXEC \n [audit].[ListarUsuarios] @Id = 1",
		},
		{
			name:          "replace with block comment before identifier",
			query:         "EXEC /* comentario */ dbo.ObtenerUsuarios @Id = 1",
			newIdentifier: "ListarUsuarios",
			expected:      "EXEC  ListarUsuarios @Id = 1",
		},
		{
			name:          "replace with comments between identifier parts",
			query:         "EXEC [dbo] /* x */ . /* y */ [ObtenerUsuarios] @Id = 1",
			newIdentifier: "[audit].[ListarUsuarios]",
			expected:      "EXEC [audit].[ListarUsuarios] @Id = 1",
		},
		{
			name:          "replace preserves params and comments after identifier are removed by sanitizer",
			query:         "EXEC dbo.ObtenerUsuarios /* comentario */ @Id = 1",
			newIdentifier: "[audit].[ListarUsuarios]",
			expected:      "EXEC [audit].[ListarUsuarios]  @Id = 1",
		},
		{
			name:          "replace lowercase exec",
			query:         "exec dbo.ObtenerUsuarios @Id = 1",
			newIdentifier: "dbo.ListarUsuarios",
			expected:      "exec dbo.ListarUsuarios @Id = 1",
		},
		{
			name:          "replace from commented fake exec to real exec",
			query:         "-- EXEC dbo.Fake\nEXEC dbo.Real @Id = 1",
			newIdentifier: "dbo.NewReal",
			expected:      "\nEXEC dbo.NewReal @Id = 1",
		},
		{
			name:          "comment only does not replace anything",
			query:         "-- EXEC dbo.ObtenerUsuarios",
			newIdentifier: "dbo.ListarUsuarios",
			expectedErr:   ErrStoredProcedureNotFound,
		},
		{
			name:          "block comment only does not replace anything",
			query:         "/* EXEC dbo.ObtenerUsuarios */",
			newIdentifier: "dbo.ListarUsuarios",
			expectedErr:   ErrStoredProcedureNotFound,
		},
		{
			name:          "invalid query select",
			query:         "SELECT * FROM Usuarios",
			newIdentifier: "dbo.ListarUsuarios",
			expectedErr:   ErrStoredProcedureNotFound,
		},
		{
			name:          "invalid query empty",
			query:         "",
			newIdentifier: "dbo.ListarUsuarios",
			expectedErr:   ErrStoredProcedureNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ReplaceStoredProcedureIdentifierInQuery(tt.query, tt.newIdentifier)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if actual != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}
