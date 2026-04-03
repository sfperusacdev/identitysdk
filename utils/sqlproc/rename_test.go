package sqlproc

import (
	"regexp"
	"strings"
	"testing"
)

type replaceProcedureNameCase struct {
	name               string
	input              string
	newName            string
	wantErr            bool
	expectedHeader     string
	oldHeaderFragments []string
	bodyFragments      []string
}

func TestReplaceProcedureName(t *testing.T) {
	cases := []replaceProcedureNameCase{
		{
			name: "simple_unqualified_name",
			input: `
CREATE PROCEDURE MyProc
AS
BEGIN
	SELECT 1
END
`,
			newName:            "RenamedProc",
			expectedHeader:     "CREATE PROCEDURE RenamedProc",
			oldHeaderFragments: []string{"CREATE PROCEDURE MyProc"},
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
			newName:            "RenamedProc",
			expectedHeader:     "CREATE PROCEDURE RenamedProc",
			oldHeaderFragments: []string{"CREATE PROCEDURE dbo.MyProc"},
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
			newName:            "RenamedProc",
			expectedHeader:     "CREATE PROCEDURE RenamedProc",
			oldHeaderFragments: []string{"CREATE PROCEDURE [dbo].[MyProc]"},
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
			newName:            "RenamedProc",
			expectedHeader:     "CREATE PROCEDURE RenamedProc",
			oldHeaderFragments: []string{"CREATE PROCEDURE [MyProc]"},
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
			newName:            "RenamedProc",
			expectedHeader:     "CREATE PROCEDURE RenamedProc",
			oldHeaderFragments: []string{"CREATE PROCEDURE _MyProc"},
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
			newName:            "RenamedProc",
			expectedHeader:     "CREATE PROCEDURE RenamedProc",
			oldHeaderFragments: []string{"CREATE PROCEDURE Proc2026"},
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
			newName:            "RenamedProc",
			expectedHeader:     "CREATE PROCEDURE RenamedProc",
			oldHeaderFragments: []string{"CREATE PROCEDURE Proc_2026_V2"},
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
			newName:            "RenamedProc",
			expectedHeader:     "CREATE OR ALTER PROC RenamedProc",
			oldHeaderFragments: []string{"CREATE OR ALTER PROC dbo.TestProc"},
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
			newName:            "RenamedProc",
			expectedHeader:     "ALTER PROCEDURE RenamedProc",
			oldHeaderFragments: []string{"ALTER PROCEDURE dbo.TestProc"},
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
			newName:            "RenamedProc",
			expectedHeader:     "cReAtE pRoCeDuRe RenamedProc",
			oldHeaderFragments: []string{"cReAtE pRoCeDuRe dbo.TestProc"},
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
			newName:            "RenamedProc",
			expectedHeader:     "CREATE PROCEDURE RenamedProc",
			oldHeaderFragments: []string{"CREATE PROCEDURE dbo   .   TestProc"},
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
			newName:            "RenamedProc",
			expectedHeader:     "CREATE PROCEDURE RenamedProc",
			oldHeaderFragments: []string{"CREATE PROCEDURE [sales reporting].[monthly summary]"},
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
			newName:            "RenamedProc",
			expectedHeader:     "CREATE PROCEDURE RenamedProc",
			oldHeaderFragments: []string{"CREATE PROCEDURE [select].[from]"},
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
			newName:            "RenamedProc",
			expectedHeader:     "CREATE PROCEDURE RenamedProc",
			oldHeaderFragments: []string{"CREATE PROCEDURE dbo.TestProc"},
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
			newName:            "RenamedProc",
			expectedHeader:     "CREATE\n/* a */\nPROCEDURE\n-- b\nRenamedProc",
			oldHeaderFragments: []string{"dbo.TestProc"},
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
			newName:            "RenamedProc",
			expectedHeader:     "CREATE PROCEDURE RenamedProc",
			oldHeaderFragments: []string{"CREATE PROCEDURE dbo.TestProc"},
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
			name: "rejects_whitespace_new_name",
			input: `
CREATE PROCEDURE TestProc
AS
BEGIN
	SELECT 1
END
`,
			newName: "   ",
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
			got, err := ReplaceProcedureName(tc.input, tc.newName)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !strings.Contains(got, tc.expectedHeader) {
				t.Fatalf("expected definition to contain header %q, got:\n%s", tc.expectedHeader, got)
			}

			for _, oldFragment := range tc.oldHeaderFragments {
				if strings.Contains(got, oldFragment) {
					t.Fatalf("expected old header fragment %q to be replaced, got:\n%s", oldFragment, got)
				}
			}

			for _, fragment := range tc.bodyFragments {
				if !strings.Contains(got, fragment) {
					t.Fatalf("expected body fragment %q to remain unchanged, got:\n%s", fragment, got)
				}
			}

			if err := ValidateProcedureDefinition(got); err != nil {
				t.Fatalf("expected renamed definition to remain valid, got %v\nDefinition:\n%s", err, got)
			}
		})
	}
}

func TestRenameProcedureDefinition(t *testing.T) {
	input := `
CREATE PROCEDURE dbo.TestProc
AS
BEGIN
	SELECT 1
END
`

	t.Run("generates_expected_name_format", func(t *testing.T) {
		newDef, newName, err := RenameProcedureDefinition(input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		matched, err := regexp.MatchString(`^p_[a-fA-F0-9]{32}$`, newName)
		if err != nil {
			t.Fatalf("unexpected regex error: %v", err)
		}
		if !matched {
			t.Fatalf("unexpected generated name: %s", newName)
		}

		if !strings.Contains(newDef, "CREATE PROCEDURE "+newName) {
			t.Fatalf("expected definition to contain generated name, got:\n%s", newDef)
		}

		if err := ValidateProcedureDefinition(newDef); err != nil {
			t.Fatalf("expected renamed definition to remain valid, got %v", err)
		}
	})

	t.Run("generates_different_names", func(t *testing.T) {
		_, name1, err := RenameProcedureDefinition(input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, name2, err := RenameProcedureDefinition(input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if name1 == name2 {
			t.Fatalf("expected different names, got same value %q", name1)
		}
	})
}
