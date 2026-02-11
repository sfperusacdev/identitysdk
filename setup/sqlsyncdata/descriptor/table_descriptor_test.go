package descriptor

import (
	"testing"
)

func TestBuildCreateTableStatement(t *testing.T) {
	tests := []struct {
		name    string
		td      TableDescriptor
		cols    []TableColumn
		wantSQL string
	}{
		{
			name: "composite pk and allowed columns",
			td: TableDescriptor{
				Table:   "orders",
				Columns: []string{"name"},
			},
			cols: []TableColumn{
				{ColumnName: "tenant_id", ColumnType: "text", ColumnNotNull: "not null", Contype: "primary key"},
				{ColumnName: "order_id", ColumnType: "integer", ColumnNotNull: "not null", Contype: "primary key"},
				{ColumnName: "name", ColumnType: "character varying(50)", ColumnNotNull: "null"},
				{ColumnName: "ignored", ColumnType: "text", ColumnNotNull: "null"},
			},
			wantSQL: "CREATE TABLE IF NOT EXISTS orders(tenant_id text not null, order_id integer not null, name TEXT, PRIMARY KEY(tenant_id, order_id))",
		},
		{
			name: "single pk only defaults",
			td: TableDescriptor{
				Table:   "audit",
				Columns: []string{},
			},
			cols: []TableColumn{
				{ColumnName: "id", ColumnType: "integer", ColumnNotNull: "not null", Contype: "primary key"},
				{ColumnName: "created_at", ColumnType: "timestamp", ColumnNotNull: "null"},
				{ColumnName: "noise", ColumnType: "text", ColumnNotNull: "null"},
			},
			wantSQL: "CREATE TABLE IF NOT EXISTS audit(id integer not null, created_at timestamp, PRIMARY KEY(id))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := tt.td.BuildCreateTableStatement(tt.cols)

			if sql != tt.wantSQL {
				t.Fatalf("sql mismatch\n got: %s\nwant: %s", sql, tt.wantSQL)
			}
		})
	}
}

func TestBuildSelectStatement(t *testing.T) {
	tests := []struct {
		name string
		td   TableDescriptor
		cols []TableColumn

		domain string
		syncAt int64

		wantSQL  string
		wantArgs []any
	}{
		{
			name: "single pk with sync",
			td: TableDescriptor{
				Table:   "orders",
				Columns: []string{"name"},
			},
			cols: []TableColumn{
				{ColumnName: "tenant_id", Contype: "primary key"},
				{ColumnName: "name"},
			},
			domain:  "acme",
			syncAt:  10,
			wantSQL: "SELECT tenant_id, name FROM orders WHERE tenant_id LIKE ? AND sync_at > ?",
			wantArgs: []any{
				"acme.%",
				int64(10),
			},
		},
		{
			name: "composite pk with skip",
			td: TableDescriptor{
				Table:                   "orders",
				Columns:                 []string{"name"},
				SkipPKPrefixCheckFilter: []string{"country_id"},
			},
			cols: []TableColumn{
				{ColumnName: "tenant_id", Contype: "primary key"},
				{ColumnName: "country_id", Contype: "primary key"},
				{ColumnName: "name"},
			},
			domain:  "corp",
			syncAt:  5,
			wantSQL: "SELECT tenant_id, country_id, name FROM orders WHERE tenant_id LIKE ? AND sync_at > ?",
			wantArgs: []any{
				"corp.%",
				int64(5),
			},
		},
		{
			name: "all pk skipped",
			td: TableDescriptor{
				Table:                   "shared",
				Columns:                 []string{"name"},
				SkipPKPrefixCheckFilter: []string{"id"},
			},
			cols: []TableColumn{
				{ColumnName: "id", Contype: "primary key"},
				{ColumnName: "name"},
			},
			domain:  "x",
			syncAt:  1,
			wantSQL: "SELECT id, name FROM shared WHERE sync_at > ?",
			wantArgs: []any{
				int64(1),
			},
		},
		{
			name: "full sync",
			td: TableDescriptor{
				Table:    "orders",
				Columns:  []string{"name"},
				FullSync: true,
			},
			cols: []TableColumn{
				{ColumnName: "tenant_id", Contype: "primary key"},
				{ColumnName: "name"},
			},
			domain:  "acme",
			syncAt:  99,
			wantSQL: "SELECT tenant_id, name FROM orders WHERE tenant_id LIKE ?",
			wantArgs: []any{
				"acme.%",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := tt.td.BuildSelectStatement(tt.cols, tt.domain, tt.syncAt)

			if sql != tt.wantSQL {
				t.Fatalf("sql mismatch\n got: %s\nwant: %s", sql, tt.wantSQL)
			}

			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args len mismatch: got %d want %d", len(args), len(tt.wantArgs))
			}

			for i := range args {
				if args[i] != tt.wantArgs[i] {
					t.Fatalf("arg[%d] mismatch: got %v want %v", i, args[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestValidateScope(t *testing.T) {
	tests := []struct {
		name        string
		td          TableDescriptor
		record      map[string]any
		primaryKeys []string
		domain      string
		wantErr     bool
	}{
		{
			name: "valido con prefijo correcto",
			td:   TableDescriptor{},
			record: map[string]any{
				"tenant_id": "acme.123",
			},
			primaryKeys: []string{"tenant_id"},
			domain:      "acme",
			wantErr:     false,
		},
		{
			name: "clave faltante",
			td:   TableDescriptor{},
			record: map[string]any{
				"otro": "acme.123",
			},
			primaryKeys: []string{"tenant_id"},
			domain:      "acme",
			wantErr:     true,
		},
		{
			name: "tipo invalido",
			td:   TableDescriptor{},
			record: map[string]any{
				"tenant_id": 123,
			},
			primaryKeys: []string{"tenant_id"},
			domain:      "acme",
			wantErr:     true,
		},
		{
			name: "prefijo incorrecto",
			td:   TableDescriptor{},
			record: map[string]any{
				"tenant_id": "otro.123",
			},
			primaryKeys: []string{"tenant_id"},
			domain:      "acme",
			wantErr:     true,
		},
		{
			name: "pk en skip no valida prefijo",
			td: TableDescriptor{
				SkipPKPrefixCheckFilter: []string{"tenant_id"},
			},
			record: map[string]any{
				"tenant_id": "otro.123",
			},
			primaryKeys: []string{"tenant_id"},
			domain:      "acme",
			wantErr:     false,
		},
		{
			name: "clave compuesta una invalida",
			td:   TableDescriptor{},
			record: map[string]any{
				"tenant_id": "acme.1",
				"id":        "otro.2",
			},
			primaryKeys: []string{"tenant_id", "id"},
			domain:      "acme",
			wantErr:     true,
		},
		{
			name: "clave compuesta valida",
			td:   TableDescriptor{},
			record: map[string]any{
				"tenant_id": "acme.1",
				"id":        "acme.2",
			},
			primaryKeys: []string{"tenant_id", "id"},
			domain:      "acme",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.td.ValidateScope(tt.record, tt.primaryKeys, tt.domain)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
