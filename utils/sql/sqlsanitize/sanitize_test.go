package sqlsanitize

import "testing"

func TestRemoveComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   \n\t  ",
			expected: "   \n\t  ",
		},
		{
			name:     "sql without comments",
			input:    "SELECT * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "line comment only",
			input:    "-- comment",
			expected: "",
		},
		{
			name:     "block comment only",
			input:    "/* comment */",
			expected: "",
		},
		{
			name:     "multiline block comment only",
			input:    "/* line 1\nline 2\nline 3 */",
			expected: "",
		},
		{
			name:     "line comment before sql",
			input:    "-- comment\nSELECT 1",
			expected: "\nSELECT 1",
		},
		{
			name:     "line comment after sql",
			input:    "SELECT 1 -- comment",
			expected: "SELECT 1 ",
		},
		{
			name:     "line comment in middle",
			input:    "SELECT 1 -- comment\nFROM dual",
			expected: "SELECT 1 \nFROM dual",
		},
		{
			name:     "block comment before sql",
			input:    "/* comment */SELECT 1",
			expected: "SELECT 1",
		},
		{
			name:     "block comment after sql",
			input:    "SELECT 1/* comment */",
			expected: "SELECT 1",
		},
		{
			name:     "block comment in middle",
			input:    "SELECT /* comment */ 1",
			expected: "SELECT  1",
		},
		{
			name:     "multiple line comments",
			input:    "-- a\n-- b\nSELECT 1 -- c\n",
			expected: "\n\nSELECT 1 \n",
		},
		{
			name:     "multiple block comments",
			input:    "/* a */SELECT/* b */ 1/* c */",
			expected: "SELECT 1",
		},
		{
			name:     "mixed line and block comments",
			input:    "/* a */SELECT 1 -- b\nFROM /* c */ dual",
			expected: "SELECT 1 \nFROM  dual",
		},
		{
			name:     "double dash inside string literal",
			input:    "SELECT 'hello -- world'",
			expected: "SELECT 'hello -- world'",
		},
		{
			name:     "block markers inside string literal",
			input:    "SELECT 'hello /* world */'",
			expected: "SELECT 'hello /* world */'",
		},
		{
			name:     "escaped single quote inside string literal",
			input:    "SELECT 'it''s ok -- still string'",
			expected: "SELECT 'it''s ok -- still string'",
		},
		{
			name:     "escaped single quote with block markers inside string literal",
			input:    "SELECT 'it''s /* not */ a comment'",
			expected: "SELECT 'it''s /* not */ a comment'",
		},
		{
			name:     "comment after string literal",
			input:    "SELECT 'abc' -- comment",
			expected: "SELECT 'abc' ",
		},
		{
			name:     "block comment after string literal",
			input:    "SELECT 'abc' /* comment */",
			expected: "SELECT 'abc' ",
		},
		{
			name:     "string literal after comment",
			input:    "-- comment\nSELECT 'abc'",
			expected: "\nSELECT 'abc'",
		},
		{
			name:     "unterminated block comment removes until end",
			input:    "SELECT 1 /* comment",
			expected: "SELECT 1 ",
		},
		{
			name:     "unterminated string literal is preserved",
			input:    "SELECT 'abc -- not comment",
			expected: "SELECT 'abc -- not comment",
		},
		{
			name:     "line comment with windows newline",
			input:    "SELECT 1 -- comment\r\nFROM dual",
			expected: "SELECT 1 \r\nFROM dual",
		},
		{
			name:     "block comment spanning windows newlines",
			input:    "SELECT /* a\r\nb\r\nc */ 1",
			expected: "SELECT  1",
		},
		{
			name:     "comments between tokens",
			input:    "SELECT/*x*/col1,/*y*/col2 FROM/*z*/table1",
			expected: "SELECTcol1,col2 FROMtable1",
		},
		{
			name:     "sql server style exec with line comment",
			input:    "EXEC -- comment\n dbo.MyProc @Id = 1",
			expected: "EXEC \n dbo.MyProc @Id = 1",
		},
		{
			name:     "sql server style exec with block comments",
			input:    "EXEC [dbo] /* x */ . /* y */ [MyProc] @Id = 1",
			expected: "EXEC [dbo]  .  [MyProc] @Id = 1",
		},
		{
			name:     "nested block comment markers are not truly nested",
			input:    "SELECT /* outer /* inner */ still here */ 1",
			expected: "SELECT  still here */ 1",
		},
		{
			name:     "dash not starting comment",
			input:    "SELECT 1 - 2",
			expected: "SELECT 1 - 2",
		},
		{
			name:     "slash star not comment when separated",
			input:    "SELECT 1 / * 2",
			expected: "SELECT 1 / * 2",
		},
		{
			name:     "adjacent string literals",
			input:    "SELECT 'a''b' + 'c--d' + 'e/*f*/g'",
			expected: "SELECT 'a''b' + 'c--d' + 'e/*f*/g'",
		},
		{
			name:     "line comment after block comment",
			input:    "/* a */-- b\nSELECT 1",
			expected: "\nSELECT 1",
		},
		{
			name:     "block comment after line comment line",
			input:    "-- a\n/* b */SELECT 1",
			expected: "\nSELECT 1",
		},
		{
			name:     "comment markers inside bracketed identifier are treated as comments by current implementation",
			input:    "SELECT [col--name], [other/*x*/col] FROM [table]",
			expected: "SELECT [col",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := RemoveComments(tt.input)
			if actual != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}
