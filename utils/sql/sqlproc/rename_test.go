package sqlproc

import (
	"strings"
	"testing"
)

type renameProcedureCase struct {
	name           string
	input          string
	newName        string
	wantErr        bool
	expectedHeader string
	bodyFragments  []string
}

func TestRenameProcedure(t *testing.T) {
	cases := []renameProcedureCase{
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
		},
		{
			name: "preserves_body_literals_and_object_id",
			input: `
CREATE PROCEDURE dbo.TestProc
AS
BEGIN
	SELECT 'dbo.TestProc'
	SELECT 'TestProc'
	SELECT OBJECT_ID('dbo.TestProc')
END
`,
			newName:        "RenamedProc",
			expectedHeader: "CREATE PROCEDURE RenamedProc",
			bodyFragments: []string{
				"'dbo.TestProc'",
				"'TestProc'",
				"OBJECT_ID('dbo.TestProc')",
			},
		},
		{
			name:    "rejects_empty_input",
			input:   "",
			newName: "RenamedProc",
			wantErr: true,
		},
		{
			name: "rejects_empty_new_name",
			input: `
CREATE PROCEDURE TestProc
AS
BEGIN
	SELECT 1
END
`,
			newName: "",
			wantErr: true,
		},
		{
			name: "rejects_invalid_definition",
			input: `
CREATE PROCEDURE TestProc
AS
BEGIN
	SELECT 1
`,
			newName: "RenamedProc",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := RenameProcedure(tc.input, tc.newName)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !strings.Contains(got.SqlDefinition, tc.expectedHeader) {
				t.Fatalf("expected definition to contain header %q, got:\n%s", tc.expectedHeader, got)
			}

			for _, fragment := range tc.bodyFragments {
				if !strings.Contains(got.SqlDefinition, fragment) {
					t.Fatalf("expected body fragment %q to remain unchanged, got:\n%s", fragment, got)
				}
			}

			if err := ValidateProcedureDefinition(got.SqlDefinition); err != nil {
				t.Fatalf("expected renamed definition to remain valid, got %v\nDefinition:\n%s", err, got)
			}
		})
	}
}

func TestCreateProcedureDefinitionVariant(t *testing.T) {
	input := `
CREATE PROCEDURE dbo.TestProc
AS
BEGIN
	SELECT 1
END
`

	t.Run("returns_original_name_variant_name_and_valid_definition", func(t *testing.T) {
		result, err := RenameProcedureWithRandomName(input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !strings.HasPrefix(result.Name, "[dbo].[p_") {
			t.Fatalf("expected variant name to start with %q, got %q", "p_", result.Name)
		}

		if !strings.Contains(result.SqlDefinition, "CREATE PROCEDURE "+result.Name) {
			t.Fatalf("expected definition to contain variant name, got:\n%s", result.SqlDefinition)
		}

		if err := ValidateProcedureDefinition(result.SqlDefinition); err != nil {
			t.Fatalf("expected variant definition to remain valid, got %v", err)
		}
	})

	t.Run("generates_different_variant_names", func(t *testing.T) {
		result1, err := RenameProcedureWithRandomName(input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		result2, err := RenameProcedureWithRandomName(input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result1.Name == result2.Name {
			t.Fatalf("expected different variant names, got same value %q", result1.Name)
		}
	})
}

func TestExtractProcedureName(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name: "unqualified_name",
			input: `
CREATE PROCEDURE MyProc
AS
BEGIN
	SELECT 1
END
`,
			expected: "MyProc",
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
			expected: "dbo.MyProc",
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
			expected: "[dbo].[MyProc]",
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
			expected: "dbo   .   TestProc",
		},
		{
			name:    "empty_input",
			input:   "",
			wantErr: true,
		},
		{
			name: "invalid_definition",
			input: `
CREATE PROCEDURE TestProc
AS
BEGIN
	SELECT 1
`,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ExtractProcedureName(tc.input)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if got != tc.expected {
				t.Fatalf("expected procedure name %q, got %q", tc.expected, got)
			}
		})
	}
}
