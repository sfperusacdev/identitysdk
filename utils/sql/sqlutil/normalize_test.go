package sqlutil

import "testing"

func TestNormalizeSQLServerIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "single_part",
			input: "Clientes",
			want:  "[Clientes]",
		},
		{
			name:  "two_parts",
			input: "dbo.Clientes",
			want:  "[dbo].[Clientes]",
		},
		{
			name:  "three_parts",
			input: "tempdb.dbo.Clientes",
			want:  "[tempdb].[dbo].[Clientes]",
		},
		{
			name:  "already_bracketed",
			input: "[dbo].[Clientes]",
			want:  "[dbo].[Clientes]",
		},
		{
			name:  "mixed_bracketed",
			input: "[dbo].Clientes",
			want:  "[dbo].[Clientes]",
		},
		{
			name:  "spaces_around_parts",
			input: " dbo . Clientes ",
			want:  "[dbo].[Clientes]",
		},
		{
			name:  "temp_table",
			input: "#tmp",
			want:  "[#tmp]",
		},
		{
			name:  "global_temp_table",
			input: "##tmp",
			want:  "[##tmp]",
		},
		{
			name:  "identifier_with_space",
			input: "dbo.Order Details",
			want:  "[dbo].[Order Details]",
		},
		{
			name:  "identifier_with_dot_inside_brackets",
			input: "[dbo].[Order.Details]",
			want:  "[dbo].[Order.Details]",
		},
		{
			name:  "escaped_closing_bracket",
			input: "[abc]]def]",
			want:  "[abc]]def]",
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only_spaces",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "trailing_dot",
			input:   "dbo.",
			wantErr: true,
		},
		{
			name:    "leading_dot",
			input:   ".Clientes",
			wantErr: true,
		},
		{
			name:    "double_dot",
			input:   "dbo..Clientes",
			wantErr: true,
		},
		{
			name:    "unterminated_bracket",
			input:   "[dbo.Clientes",
			wantErr: true,
		},
		{
			name:    "unexpected_closing_bracket",
			input:   "dbo].Clientes",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeSQLServerIdentifier(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tt.input, err)
			}

			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
