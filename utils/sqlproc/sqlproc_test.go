package sqlproc

import "testing"

func TestValidateProcedureDefinition(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid procedure",
			input: `
CREATE PROCEDURE dbo.GetUsers
AS
BEGIN
	SELECT 1
END
`,
			wantErr: false,
		},
		{
			name: "missing as",
			input: `
CREATE PROCEDURE dbo.GetUsers
BEGIN
	SELECT 1
END
`,
			wantErr: true,
		},
		{
			name: "missing begin",
			input: `
CREATE PROCEDURE dbo.GetUsers
AS
	SELECT 1
`,
			wantErr: true,
		},
		{
			name: "extra content after end",
			input: `
CREATE PROCEDURE dbo.GetUsers
AS
BEGIN
	SELECT 1
END
SELECT 2
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProcedureDefinition(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateProcedureDefinition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
