package sqlproc

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceProcedureName_WithVariousValidProcedureNames(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		newName         string
		expectedHeader  string
		oldHeaderPieces []string
	}{
		{
			name: "simple_unqualified_name",
			input: `
CREATE PROCEDURE MyProc
AS
BEGIN
	SELECT 1
END
`,
			newName:        "RenamedProc",
			expectedHeader: "CREATE PROCEDURE RenamedProc",
			oldHeaderPieces: []string{
				"CREATE PROCEDURE MyProc",
			},
		},
		{
			name: "schema_qualified_name",
			input: `
CREATE PROCEDURE dbo.MyProc
AS
BEGIN
	SELECT 1
END
`,
			newName:        "RenamedProc",
			expectedHeader: "CREATE PROCEDURE RenamedProc",
			oldHeaderPieces: []string{
				"CREATE PROCEDURE dbo.MyProc",
			},
		},
		{
			name: "bracketed_schema_and_name",
			input: `
CREATE PROCEDURE [dbo].[MyProc]
AS
BEGIN
	SELECT 1
END
`,
			newName:        "RenamedProc",
			expectedHeader: "CREATE PROCEDURE RenamedProc",
			oldHeaderPieces: []string{
				"CREATE PROCEDURE [dbo].[MyProc]",
			},
		},
		{
			name: "bracketed_name_only",
			input: `
CREATE PROCEDURE [MyProc]
AS
BEGIN
	SELECT 1
END
`,
			newName:        "RenamedProc",
			expectedHeader: "CREATE PROCEDURE RenamedProc",
			oldHeaderPieces: []string{
				"CREATE PROCEDURE [MyProc]",
			},
		},
		{
			name: "underscore_prefix",
			input: `
CREATE PROCEDURE _MyProc
AS
BEGIN
	SELECT 1
END
`,
			newName:        "RenamedProc",
			expectedHeader: "CREATE PROCEDURE RenamedProc",
			oldHeaderPieces: []string{
				"CREATE PROCEDURE _MyProc",
			},
		},
		{
			name: "name_with_digits",
			input: `
CREATE PROCEDURE Proc2026
AS
BEGIN
	SELECT 1
END
`,
			newName:        "RenamedProc",
			expectedHeader: "CREATE PROCEDURE RenamedProc",
			oldHeaderPieces: []string{
				"CREATE PROCEDURE Proc2026",
			},
		},
		{
			name: "name_with_underscores_and_digits",
			input: `
CREATE PROCEDURE Proc_2026_V2
AS
BEGIN
	SELECT 1
END
`,
			newName:        "RenamedProc",
			expectedHeader: "CREATE PROCEDURE RenamedProc",
			oldHeaderPieces: []string{
				"CREATE PROCEDURE Proc_2026_V2",
			},
		},
		{
			name: "create_or_alter_proc",
			input: `
CREATE OR ALTER PROC dbo.TestProc
AS
BEGIN
	SELECT 1
END
`,
			newName:        "RenamedProc",
			expectedHeader: "CREATE OR ALTER PROC RenamedProc",
			oldHeaderPieces: []string{
				"CREATE OR ALTER PROC dbo.TestProc",
			},
		},
		{
			name: "alter_procedure",
			input: `
ALTER PROCEDURE dbo.TestProc
AS
BEGIN
	SELECT 1
END
`,
			newName:        "RenamedProc",
			expectedHeader: "ALTER PROCEDURE RenamedProc",
			oldHeaderPieces: []string{
				"ALTER PROCEDURE dbo.TestProc",
			},
		},
		{
			name: "mixed_casing",
			input: `
cReAtE pRoCeDuRe dbo.TestProc
AS
BEGIN
	SELECT 1
END
`,
			newName:        "RenamedProc",
			expectedHeader: "cReAtE pRoCeDuRe RenamedProc",
			oldHeaderPieces: []string{
				"cReAtE pRoCeDuRe dbo.TestProc",
			},
		},
		{
			name: "spaces_around_dot",
			input: `
CREATE PROCEDURE dbo   .   TestProc
AS
BEGIN
	SELECT 1
END
`,
			newName:        "RenamedProc",
			expectedHeader: "CREATE PROCEDURE RenamedProc",
			oldHeaderPieces: []string{
				"CREATE PROCEDURE dbo   .   TestProc",
			},
		},
		{
			name: "bracketed_identifiers_with_spaces",
			input: `
CREATE PROCEDURE [sales reporting].[monthly summary]
AS
BEGIN
	SELECT 1
END
`,
			newName:        "RenamedProc",
			expectedHeader: "CREATE PROCEDURE RenamedProc",
			oldHeaderPieces: []string{
				"CREATE PROCEDURE [sales reporting].[monthly summary]",
			},
		},
		{
			name: "bracketed_reserved_words",
			input: `
CREATE PROCEDURE [select].[from]
AS
BEGIN
	SELECT 1
END
`,
			newName:        "RenamedProc",
			expectedHeader: "CREATE PROCEDURE RenamedProc",
			oldHeaderPieces: []string{
				"CREATE PROCEDURE [select].[from]",
			},
		},
		{
			name: "parameters_after_procedure_name",
			input: `
CREATE PROCEDURE dbo.TestProc
	@param1 INT,
	@param2 NVARCHAR(50)
AS
BEGIN
	SELECT @param1, @param2
END
`,
			newName:        "RenamedProc",
			expectedHeader: "CREATE PROCEDURE RenamedProc",
			oldHeaderPieces: []string{
				"CREATE PROCEDURE dbo.TestProc",
			},
		},
		{
			name: "comments_between_tokens",
			input: `
CREATE
/* a */
PROCEDURE
-- b
dbo.TestProc
AS
BEGIN
	SELECT 1
END
`,
			newName:        "RenamedProc",
			expectedHeader: "CREATE\n/* a */\nPROCEDURE\n-- b\nRenamedProc",
			oldHeaderPieces: []string{
				"dbo.TestProc",
			},
		},
		{
			name: "body_contains_old_procedure_name_in_string_literal",
			input: `
CREATE PROCEDURE dbo.TestProc
AS
BEGIN
	SELECT 'dbo.TestProc'
	SELECT 'TestProc'
END
`,
			newName:        "RenamedProc",
			expectedHeader: "CREATE PROCEDURE RenamedProc",
			oldHeaderPieces: []string{
				"CREATE PROCEDURE dbo.TestProc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplaceProcedureName(tt.input, tt.newName)

			require.NoError(t, err)
			require.Contains(t, got, tt.expectedHeader)

			for _, oldPiece := range tt.oldHeaderPieces {
				require.NotContains(t, got, oldPiece)
			}

			require.NoError(t, ValidateProcedureDefinition(got))
		})
	}
}

func TestReplaceProcedureName_PreservesBodyContent(t *testing.T) {
	input := `
CREATE PROCEDURE dbo.TestProc
AS
BEGIN
	SELECT 'dbo.TestProc'
	SELECT 'TestProc'
	SELECT OBJECT_ID('dbo.TestProc')
END
`

	got, err := ReplaceProcedureName(input, "RenamedProc")

	require.NoError(t, err)
	require.Contains(t, got, "'dbo.TestProc'")
	require.Contains(t, got, "'TestProc'")
	require.Contains(t, got, "OBJECT_ID('dbo.TestProc')")
	require.Contains(t, got, "CREATE PROCEDURE RenamedProc")
}

func TestReplaceProcedureName_RejectsEmptyInput(t *testing.T) {
	_, err := ReplaceProcedureName("", "RenamedProc")
	require.Error(t, err)
}

func TestReplaceProcedureName_RejectsEmptyNewName(t *testing.T) {
	input := `
CREATE PROCEDURE TestProc
AS
BEGIN
	SELECT 1
END
`

	_, err := ReplaceProcedureName(input, "")
	require.Error(t, err)
}

func TestReplaceProcedureName_RejectsWhitespaceNewName(t *testing.T) {
	input := `
CREATE PROCEDURE TestProc
AS
BEGIN
	SELECT 1
END
`

	_, err := ReplaceProcedureName(input, "   ")
	require.Error(t, err)
}

func TestReplaceProcedureName_RejectsInvalidProcedureDefinition(t *testing.T) {
	input := `
CREATE PROCEDURE TestProc
AS
BEGIN
	SELECT 1
`

	_, err := ReplaceProcedureName(input, "RenamedProc")
	require.Error(t, err)
}

func TestRenameProcedureDefinition_GeneratesExpectedNameFormat(t *testing.T) {
	input := `
CREATE PROCEDURE dbo.TestProc
AS
BEGIN
	SELECT 1
END
`

	newDef, newName, err := RenameProcedureDefinition(input)

	require.NoError(t, err)
	require.Regexp(t, regexp.MustCompile(`^p_[a-fA-F0-9]{32}$`), newName)
	require.Contains(t, newDef, "CREATE PROCEDURE "+newName)
	require.NoError(t, ValidateProcedureDefinition(newDef))
}

func TestRenameProcedureDefinition_GeneratesDifferentNames(t *testing.T) {
	input := `
CREATE PROCEDURE dbo.TestProc
AS
BEGIN
	SELECT 1
END
`

	_, name1, err := RenameProcedureDefinition(input)
	require.NoError(t, err)

	_, name2, err := RenameProcedureDefinition(input)
	require.NoError(t, err)

	require.NotEqual(t, name1, name2)
}

func TestReplaceProcedureName_WithManyValidProcedureNames(t *testing.T) {
	validNames := []string{
		"MyProc",
		"_MyProc",
		"Proc2026",
		"Proc_2026_V2",
		"[MyProc]",
		"dbo.MyProc",
		"[dbo].[MyProc]",
		"dbo   .   MyProc",
		"[sales reporting].[monthly summary]",
		"[select].[from]",
	}

	for _, procName := range validNames {
		t.Run(strings.ReplaceAll(procName, " ", "_"), func(t *testing.T) {
			input := `
CREATE PROCEDURE ` + procName + `
AS
BEGIN
	SELECT 1
END
`

			got, err := ReplaceProcedureName(input, "RenamedProc")

			require.NoError(t, err)
			require.Contains(t, got, "RenamedProc")
			require.NotContains(t, got, "CREATE PROCEDURE "+procName)
			require.NoError(t, ValidateProcedureDefinition(got))
		})
	}
}
