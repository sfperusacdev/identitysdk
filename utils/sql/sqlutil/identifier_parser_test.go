package sqlutil

import "testing"

func TestParseSQLServerIdentifier(t *testing.T) {
	cases := []struct {
		name           string
		input          string
		expectedObject string
		expectedSchema []string
		expectedString string
		wantErr        bool
	}{
		{
			name:           "single_part_name",
			input:          "MyTable",
			expectedObject: "MyTable",
			expectedString: "[MyTable]",
		},
		{
			name:           "schema_and_object",
			input:          "dbo.MyTable",
			expectedObject: "MyTable",
			expectedSchema: []string{"dbo"},
			expectedString: "[dbo].[MyTable]",
		},
		{
			name:           "database_schema_object",
			input:          "MyDb.dbo.MyTable",
			expectedObject: "MyTable",
			expectedSchema: []string{"MyDb", "dbo"},
			expectedString: "[MyDb].[dbo].[MyTable]",
		},
		{
			name:           "server_database_schema_object",
			input:          "MyServer.MyDb.dbo.MyTable",
			expectedObject: "MyTable",
			expectedSchema: []string{"MyServer", "MyDb", "dbo"},
			expectedString: "[MyServer].[MyDb].[dbo].[MyTable]",
		},
		{
			name:           "bracketed_identifier",
			input:          "[dbo].[MyTable]",
			expectedObject: "MyTable",
			expectedSchema: []string{"dbo"},
			expectedString: "[dbo].[MyTable]",
		},
		{
			name:           "bracketed_identifier_with_spaces",
			input:          "[sales reporting].[monthly summary]",
			expectedObject: "monthly summary",
			expectedSchema: []string{"sales reporting"},
			expectedString: "[sales reporting].[monthly summary]",
		},
		{
			name:           "spaces_around_dot",
			input:          "dbo   .   MyTable",
			expectedObject: "MyTable",
			expectedSchema: []string{"dbo"},
			expectedString: "[dbo].[MyTable]",
		},
		{
			name:           "escaped_closing_bracket",
			input:          "[dbo].[My]]Table]",
			expectedObject: "My]Table",
			expectedSchema: []string{"dbo"},
			expectedString: "[dbo].[My]]Table]",
		},
		{
			name:    "empty_input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace_input",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "empty_part_between_dots",
			input:   "dbo..MyTable",
			wantErr: true,
		},
		{
			name:    "trailing_dot",
			input:   "dbo.MyTable.",
			wantErr: true,
		},
		{
			name:    "unterminated_bracket",
			input:   "[dbo].[MyTable",
			wantErr: true,
		},
		{
			name:    "unexpected_closing_bracket",
			input:   "dbo].MyTable",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseSQLServerIdentifier(tc.input)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if got.ObjectName != tc.expectedObject {
				t.Fatalf("expected object name %q, got %q", tc.expectedObject, got.ObjectName)
			}

			if len(got.SchemaPath) != len(tc.expectedSchema) {
				t.Fatalf("expected schema path length %d, got %d (%v)", len(tc.expectedSchema), len(got.SchemaPath), got.SchemaPath)
			}

			for i := range tc.expectedSchema {
				if got.SchemaPath[i] != tc.expectedSchema[i] {
					t.Fatalf("expected schema path %v, got %v", tc.expectedSchema, got.SchemaPath)
				}
			}

			if got.String() != tc.expectedString {
				t.Fatalf("expected string %q, got %q", tc.expectedString, got.String())
			}
		})
	}
}
