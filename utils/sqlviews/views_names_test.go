package sqlviews

import (
	"reflect"
	"testing"
)

func TestFindViewNames_Postgres_AllCases(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected []string
	}{
		{
			name:     "single simple",
			sql:      `create view vista1 as select 1;`,
			expected: []string{"vista1"},
		},
		{
			name:     "uppercase",
			sql:      `CREATE VIEW vista2 AS SELECT 1;`,
			expected: []string{"vista2"},
		},
		{
			name:     "schema qualified",
			sql:      `create view public.vista3 as select 1;`,
			expected: []string{"public.vista3"},
		},
		{
			name:     "quoted view",
			sql:      `create view public."vista rara" as select 1;`,
			expected: []string{`public."vistarara"`},
		},
		{
			name:     "quoted schema",
			sql:      `create view "mi schema".vista4 as select 1;`,
			expected: []string{`"mischema".vista4`},
		},
		{
			name:     "quoted both",
			sql:      `create view "mi schema"."vista rara" as select 1;`,
			expected: []string{`"mischema"."vistarara"`},
		},
		{
			name:     "or replace",
			sql:      `create or replace view vista5 as select 1;`,
			expected: []string{"vista5"},
		},
		{
			name: "with newlines",
			sql: `
				create
				   or    replace
				view
				   vista6
				as
				select 1;
			`,
			expected: []string{"vista6"},
		},
		{
			name:     "spaces around dot",
			sql:      `create view public   .   vista7 as select 1;`,
			expected: []string{"public.vista7"},
		},
		{
			name:     "quoted with spaces around dot",
			sql:      `create view "mi schema"   .   "otra vista" as select 1;`,
			expected: []string{`"mischema"."otravista"`},
		},
		{
			name:     "tabs and weird spacing",
			sql:      "create\tview\tpublic\t.\tvista8\tas\tselect 1;",
			expected: []string{"public.vista8"},
		},
		{
			name: "multiple views mixed",
			sql: `
				create view v1 as select 1;

				create or replace view public.v2 as select 2;

				create
				view
				"vista rara"
				as select 3;

				create view "mi schema".v4 as select 4;
			`,
			expected: []string{
				"v1",
				"public.v2",
				`"vistarara"`,
				`"mischema".v4`,
			},
		},
		{
			name: "no views",
			sql: `
				select * from tabla;
				drop table test;
			`,
			expected: []string{},
		},
		{
			name: "complex formatting",
			sql: `
				CREATE
				     VIEW
				         public
				             .
				         "vista con    espacios"
				AS
				SELECT 1;

				CREATE OR
				REPLACE
				VIEW
				    "otro schema"
				        .
				    otra_vista
				AS
				SELECT 2;
			`,
			expected: []string{
				`public."vistaconespacios"`,
				`"otroschema".otra_vista`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindViewNames(tt.sql)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("expected %v got %v", tt.expected, got)
			}
		})
	}
}
