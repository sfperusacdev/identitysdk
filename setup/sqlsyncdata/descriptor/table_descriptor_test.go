package descriptor

import (
	"strings"
	"testing"
)

func TestBuildCreateTableStatement_ColumnFiltering(t *testing.T) {
	td := TableDescriptor{
		Table:   "users",
		Columns: []string{"email", "age"},
	}

	cols := []TableColumn{
		{ColumnName: "id", ColumnType: "integer", ColumnNotNull: "not null", Contype: "primary key"},
		{ColumnName: "email", ColumnType: "text", ColumnNotNull: "not null"},
		{ColumnName: "age", ColumnType: "integer", ColumnNotNull: "null"},
		{ColumnName: "password", ColumnType: "text", ColumnNotNull: "not null"},
	}

	sql := td.BuildCreateTableStatement(cols)

	if !strings.Contains(sql, "email text not null") {
		t.Fatalf("expected email column, got: %s", sql)
	}
	if !strings.Contains(sql, "age integer") {
		t.Fatalf("expected age column, got: %s", sql)
	}
	if strings.Contains(sql, "password") {
		t.Fatalf("did not expect password column, got: %s", sql)
	}
}

func TestBuildCreateTableStatement_DefaultColumnsOnly(t *testing.T) {
	td := TableDescriptor{
		Table:   "logs",
		Columns: []string{},
	}

	cols := []TableColumn{
		{ColumnName: "id", ColumnType: "integer", ColumnNotNull: "not null", Contype: "primary key"},
		{ColumnName: "created_at", ColumnType: "timestamp", ColumnNotNull: "null"},
		{ColumnName: "updated_at", ColumnType: "timestamp", ColumnNotNull: "null"},
		{ColumnName: "noise", ColumnType: "text", ColumnNotNull: "null"},
	}

	sql := td.BuildCreateTableStatement(cols)

	if strings.Contains(sql, "noise") {
		t.Fatalf("did not expect noise column, got: %s", sql)
	}
	if !strings.Contains(sql, "created_at timestamp") {
		t.Fatalf("expected created_at column, got: %s", sql)
	}
}

func TestBuildCreateTableStatement_IndexesCreated(t *testing.T) {
	td := TableDescriptor{
		Table:   "sync_table",
		Columns: []string{"sync_at", "deleted_at"},
	}

	cols := []TableColumn{
		{ColumnName: "id", ColumnType: "integer", ColumnNotNull: "not null", Contype: "primary key"},
		{ColumnName: "sync_at", ColumnType: "timestamp", ColumnNotNull: "null"},
		{ColumnName: "deleted_at", ColumnType: "timestamp", ColumnNotNull: "null"},
	}

	sql := td.BuildCreateTableStatement(cols)

	if !strings.Contains(sql, "CREATE INDEX IF NOT EXISTS idx_sync_table_sync_at") {
		t.Fatalf("expected sync_at index, got: %s", sql)
	}
	if !strings.Contains(sql, "CREATE INDEX IF NOT EXISTS idx_sync_table_deleted_at") {
		t.Fatalf("expected deleted_at index, got: %s", sql)
	}
}

func TestBuildCreateTableStatement_NoIndexesWhenColumnsMissing(t *testing.T) {
	td := TableDescriptor{
		Table:   "plain",
		Columns: []string{},
	}

	cols := []TableColumn{
		{ColumnName: "id", ColumnType: "integer", ColumnNotNull: "not null", Contype: "primary key"},
	}

	sql := td.BuildCreateTableStatement(cols)

	if strings.Contains(sql, "CREATE INDEX") {
		t.Fatalf("did not expect any index, got: %s", sql)
	}
}

func TestBuildCreateTableStatement_PrimaryKeysAlwaysIncluded(t *testing.T) {
	td := TableDescriptor{
		Table:   "mixed",
		Columns: []string{"allowed"},
	}

	cols := []TableColumn{
		{ColumnName: "pk1", ColumnType: "integer", ColumnNotNull: "not null", Contype: "primary key"},
		{ColumnName: "allowed", ColumnType: "text", ColumnNotNull: "null"},
		{ColumnName: "blocked", ColumnType: "text", ColumnNotNull: "null"},
	}

	sql := td.BuildCreateTableStatement(cols)

	if !strings.Contains(sql, "pk1 integer not null") {
		t.Fatalf("expected pk1 column, got: %s", sql)
	}
	if strings.Contains(sql, "blocked") {
		t.Fatalf("did not expect blocked column, got: %s", sql)
	}
}

func TestBuildCreateTableStatement_ConfiguredPrimaryKeys(t *testing.T) {
	td := TableDescriptor{
		Table:       "readonly_items",
		Columns:     []string{"name"},
		PrimaryKeys: []string{"code"},
	}

	cols := []TableColumn{
		{ColumnName: "code", ColumnType: "text", ColumnNotNull: "not null"},
		{ColumnName: "name", ColumnType: "text", ColumnNotNull: "null"},
		{ColumnName: "noise", ColumnType: "text", ColumnNotNull: "null"},
	}

	sql := td.BuildCreateTableStatement(cols)

	if sql != "CREATE TABLE IF NOT EXISTS readonly_items(code text not null, name text, PRIMARY KEY(code))" {
		t.Fatalf("sql mismatch: %s", sql)
	}
}

func TestTableDescriptor_PublicName(t *testing.T) {
	tests := []struct {
		name string
		td   TableDescriptor
		want string
	}{
		{
			name: "uses alias when configured",
			td:   TableDescriptor{Table: "physical_users", Alias: "client_users"},
			want: "client_users",
		},
		{
			name: "falls back to physical table",
			td:   TableDescriptor{Table: "users"},
			want: "users",
		},
		{
			name: "blank alias falls back to physical table",
			td:   TableDescriptor{Table: "users", Alias: "  "},
			want: "users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.td.PublicName(); got != tt.want {
				t.Fatalf("PublicName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeColumnType(t *testing.T) {
	tests := []struct {
		name       string
		columnType string
		want       string
	}{
		{
			name:       "character varying becomes text",
			columnType: "character varying(255)",
			want:       "TEXT",
		},
		{
			name:       "numeric becomes real",
			columnType: "numeric(10,2)",
			want:       "REAL",
		},
		{
			name:       "other type is unchanged",
			columnType: "bigint",
			want:       "bigint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeColumnType(tt.columnType); got != tt.want {
				t.Errorf("NormalizeColumnType(%q) = %q, want %q", tt.columnType, got, tt.want)
			}
		})
	}
}

func TestTableDescriptor_IsReadyOnly(t *testing.T) {
	tests := []struct {
		name         string
		descriptor   TableDescriptor
		columns      []TableColumn
		wantReadOnly bool
	}{
		{
			name:         "explicit read only",
			descriptor:   TableDescriptor{ReadOnly: true},
			columns:      []TableColumn{{ColumnName: "sync_at"}},
			wantReadOnly: true,
		},
		{
			name:         "sync column allows writes",
			descriptor:   TableDescriptor{},
			columns:      []TableColumn{{ColumnName: "sync_at"}},
			wantReadOnly: false,
		},
		{
			name:         "missing sync column is read only",
			descriptor:   TableDescriptor{},
			columns:      []TableColumn{{ColumnName: "id"}},
			wantReadOnly: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.descriptor.IsReadyOnly(tt.columns); got != tt.wantReadOnly {
				t.Errorf("IsReadyOnly() = %t, want %t", got, tt.wantReadOnly)
			}
		})
	}
}

func TestBuildCreateTableStatement_UsesAliasForClientTable(t *testing.T) {
	td := TableDescriptor{
		Table:   "physical_users",
		Alias:   "client_users",
		Columns: []string{"sync_at"},
	}
	cols := []TableColumn{
		{ColumnName: "id", ColumnType: "text", ColumnNotNull: "not null", Contype: "primary key"},
		{ColumnName: "sync_at", ColumnType: "bigint", ColumnNotNull: "not null"},
	}

	want := "CREATE TABLE IF NOT EXISTS client_users(id text not null, sync_at bigint not null, PRIMARY KEY(id)); CREATE INDEX IF NOT EXISTS idx_client_users_sync_at ON client_users(sync_at)"
	if got := td.BuildCreateTableStatement(cols); got != want {
		t.Fatalf("BuildCreateTableStatement() = %q, want %q", got, want)
	}
}

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
				Columns: []string{"name", "cliente_id"},
			},
			cols: []TableColumn{
				{ColumnName: "tenant_id", ColumnType: "text", ColumnNotNull: "not null", Contype: "primary key"},
				{ColumnName: "order_id", ColumnType: "integer", ColumnNotNull: "not null", Contype: "primary key"},
				{ColumnName: "cliente_id", ColumnType: "integer", ColumnNotNull: "not null", Contype: "unique"},
				{ColumnName: "name", ColumnType: "character varying(50)", ColumnNotNull: "null"},
				{ColumnName: "ignored", ColumnType: "text", ColumnNotNull: "null"},
			},
			wantSQL: "CREATE TABLE IF NOT EXISTS orders(tenant_id text not null, order_id integer not null, cliente_id integer not null, name TEXT, PRIMARY KEY(tenant_id, order_id))",
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
		{
			name: "sync_at and deleted_at create indexes",
			td: TableDescriptor{
				Table:   "sync_table",
				Columns: []string{"sync_at", "deleted_at"},
			},
			cols: []TableColumn{
				{ColumnName: "id", ColumnType: "integer", ColumnNotNull: "not null", Contype: "primary key"},
				{ColumnName: "sync_at", ColumnType: "timestamp", ColumnNotNull: "null"},
				{ColumnName: "deleted_at", ColumnType: "timestamp", ColumnNotNull: "null"},
			},
			wantSQL: "CREATE TABLE IF NOT EXISTS sync_table(id integer not null, sync_at timestamp, deleted_at timestamp, PRIMARY KEY(id)); CREATE INDEX IF NOT EXISTS idx_sync_table_deleted_at ON sync_table(deleted_at); CREATE INDEX IF NOT EXISTS idx_sync_table_sync_at ON sync_table(sync_at)",
		},
		{
			name: "no indexes when special columns are missing",
			td: TableDescriptor{
				Table:   "plain",
				Columns: []string{},
			},
			cols: []TableColumn{
				{ColumnName: "id", ColumnType: "integer", ColumnNotNull: "not null", Contype: "primary key"},
			},
			wantSQL: "CREATE TABLE IF NOT EXISTS plain(id integer not null, PRIMARY KEY(id))",
		},
		{
			name: "primary key included even if not allowed",
			td: TableDescriptor{
				Table:   "mixed",
				Columns: []string{"allowed"},
			},
			cols: []TableColumn{
				{ColumnName: "pk", ColumnType: "integer", ColumnNotNull: "not null", Contype: "primary key"},
				{ColumnName: "allowed", ColumnType: "text", ColumnNotNull: "null"},
				{ColumnName: "blocked", ColumnType: "text", ColumnNotNull: "null"},
			},
			wantSQL: "CREATE TABLE IF NOT EXISTS mixed(pk integer not null, allowed text, PRIMARY KEY(pk))",
		},
		{
			name: "multiple primary keys only",
			td: TableDescriptor{
				Table:   "only_pk",
				Columns: []string{},
			},
			cols: []TableColumn{
				{ColumnName: "a", ColumnType: "integer", ColumnNotNull: "not null", Contype: "primary key"},
				{ColumnName: "b", ColumnType: "integer", ColumnNotNull: "not null", Contype: "primary key"},
				{ColumnName: "noise", ColumnType: "text", ColumnNotNull: "null"},
			},
			wantSQL: "CREATE TABLE IF NOT EXISTS only_pk(a integer not null, b integer not null, PRIMARY KEY(a, b))",
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
		{
			name: "configured primary key",
			td: TableDescriptor{
				Table:       "readonly_items",
				Columns:     []string{"name"},
				PrimaryKeys: []string{"code"},
			},
			cols: []TableColumn{
				{ColumnName: "code"},
				{ColumnName: "name"},
			},
			domain:  "acme",
			syncAt:  20,
			wantSQL: "SELECT code, name FROM readonly_items WHERE code LIKE ? AND sync_at > ?",
			wantArgs: []any{
				"acme.%",
				int64(20),
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

func TestBuildSelectStatement_UsesPhysicalTableWithAlias(t *testing.T) {
	td := TableDescriptor{
		Table:   "physical_users",
		Alias:   "client_users",
		Columns: []string{"name"},
	}
	cols := []TableColumn{
		{ColumnName: "id", Contype: "primary key"},
		{ColumnName: "name"},
	}

	query, _ := td.BuildSelectStatement(cols, "acme", 0)
	if !strings.Contains(query, " FROM physical_users ") {
		t.Fatalf("expected select to target physical table, got: %s", query)
	}
	if strings.Contains(query, "client_users") {
		t.Fatalf("expected select not to use client alias, got: %s", query)
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
