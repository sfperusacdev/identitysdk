package descriptor

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

type TableDescriptor struct {
	Table string `gorm:"primaryKey"`
	// Physical name of the table in the database

	// Columns allowed for reading / synchronization
	Columns []string

	// Primary key prefixes that are exempt from the prefix validation rule %s
	SkipPKPrefixCheckFilter []string

	SinceDays uint
	// Number of days back from now to start synchronization
	// Only used when FullSync is false
	// If the value is 0, the entire historical dataset is synchronized

	// SyncDelay time.Duration
	// Time subtracted from the last synchronization checkpoint
	// Used to re-read a small overlap window and avoid missing recent changes

	FullSync bool
	// Indicates whether synchronization must be full
	// When true, it takes precedence over SinceDays
}

func (td TableDescriptor) StartSyncAt() time.Time {
	if td.SinceDays == 0 {
		return time.Unix(0, 0).UTC()

	}
	return time.Now().AddDate(0, 0, -int(td.SinceDays))
}

var defaultColumnNames = []string{
	"created_at",
	"created_by",
	"updated_at",
	"updated_by",
	"deleted_at",
	"sync_at",
}

const primaryKeyKeyword = "primary key"

type TableColumn struct {
	TableName     string
	ColumnName    string
	ColumnType    string
	ColumnNotNull string
	Contype       string
}

func (tc TableColumn) IsPrimaryKey() bool {
	return strings.ToLower(strings.TrimSpace(tc.Contype)) == primaryKeyKeyword
}

func (td TableDescriptor) BuildCreateTableStatement(tableColumns []TableColumn) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "CREATE TABLE IF NOT EXISTS %s(", td.Table)

	allowedColumns := append([]string{}, td.Columns...)
	allowedColumns = append(allowedColumns, defaultColumnNames...)

	var primaryKeys []string
	columnCount := 0

	for _, col := range tableColumns {
		columnName := strings.ToLower(col.ColumnName)
		if !col.IsPrimaryKey() && !slices.Contains(allowedColumns, columnName) {
			continue
		}
		if columnCount > 0 {
			builder.WriteString(", ")
		}
		columnCount++

		builder.WriteString(columnName)
		builder.WriteByte(' ')

		columnType := normalizeColumnType(col.ColumnType)
		builder.WriteString(columnType)

		if strings.TrimSpace(col.ColumnNotNull) != "null" {
			builder.WriteByte(' ')
			builder.WriteString(col.ColumnNotNull)
		}

		if col.IsPrimaryKey() {
			primaryKeys = append(primaryKeys, columnName)
			continue
		}

		builder.WriteString(col.Contype)
	}

	if len(primaryKeys) > 0 {
		builder.WriteString(", PRIMARY KEY(")
		builder.WriteString(strings.Join(primaryKeys, ", "))
		builder.WriteString(")")
	}

	builder.WriteByte(')')

	return builder.String()
}

func normalizeColumnType(columnType string) string {
	switch {
	case strings.HasPrefix(columnType, "character varying"):
		return "TEXT"
	case strings.HasPrefix(columnType, "numeric"):
		return "REAL"
	default:
		return columnType
	}
}

func (td TableDescriptor) BuildSelectStatement(tableColumns []TableColumn, domain string, syncAt int64) (string, []any) {
	var builder strings.Builder

	fmt.Fprintf(&builder, "SELECT ")

	allowedColumns := append([]string{}, td.Columns...)
	allowedColumns = append(allowedColumns, defaultColumnNames...)

	columnCount := 0
	var primaryKeys []string
	for _, col := range tableColumns {
		columnName := strings.ToLower(col.ColumnName)
		if !col.IsPrimaryKey() &&
			!slices.Contains(allowedColumns, columnName) {
			continue
		}

		if col.IsPrimaryKey() &&
			!slices.Contains(td.SkipPKPrefixCheckFilter, columnName) {
			primaryKeys = append(primaryKeys, columnName)
		}

		if columnCount > 0 {
			builder.WriteString(", ")
		}
		columnCount++

		builder.WriteString(columnName)
	}

	fmt.Fprintf(&builder, " FROM %s", td.Table)
	var where []string
	var whereArgs = []any{}

	if len(primaryKeys) > 0 {
		for _, p := range primaryKeys {
			where = append(where, fmt.Sprintf("%s LIKE ?", p))
			whereArgs = append(whereArgs, fmt.Sprintf("%s.%%", domain))
		}
	}

	if !td.FullSync {
		where = append(where, "sync_at > ?")
		whereArgs = append(whereArgs, syncAt)
	}

	qry := builder.String()
	if len(where) > 0 {
		qry = fmt.Sprintf("%s WHERE %s", qry, strings.Join(where, " AND "))
	}

	return qry, whereArgs
}

func (td TableDescriptor) ValidateScope(
	record map[string]any,
	primaryKeys []string,
	domain string,
) error {
	prefix := domain + "."
	for _, pk := range primaryKeys {
		pk = strings.ToLower(pk)

		if slices.Contains(td.SkipPKPrefixCheckFilter, pk) {
			continue
		}

		value, exists := record[pk]
		if !exists {
			return fmt.Errorf("falta el campo obligatorio de la clave: %s", pk)
		}

		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("el campo %s debe ser string para validar el dominio", pk)
		}

		if !strings.HasPrefix(strVal, prefix) {
			return fmt.Errorf("el valor de %s no pertenece al dominio %q", pk, domain)
		}
	}
	return nil
}
