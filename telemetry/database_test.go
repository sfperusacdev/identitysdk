package telemetry

import "testing"

func TestSQLOperation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, query, want string }{
		{"parameterized select", "SELECT * FROM workers WHERE code = $1", "SELECT"},
		{"literal secret", "select 'secret-value'", "SELECT"},
		{"update", " UPDATE workers SET password = 'secret'", "UPDATE"},
		{"unknown statement", "CALL secret_procedure('secret')", "SQL"},
		{"comment before query", "/* secret */ SELECT 1", "SQL"},
		{"empty query", "", "SQL"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := sqlOperation(tt.query); got != tt.want {
				t.Fatalf("sqlOperation() = %q, want %q", got, tt.want)
			}
		})
	}
}
