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

	Columns []string
	// Columns allowed for reading / synchronization

	Global bool
	// Indicates whether the table is global (true) or domain-specific (false)

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

func (td TableDescriptor) BuildCreateTableStatement(tableColumns []TableColumn) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "CREATE TABLE IF NOT EXISTS %s(", td.Table)

	allowedColumns := append([]string{}, td.Columns...)
	allowedColumns = append(allowedColumns, defaultColumnNames...)

	var primaryKeys []string
	columnCount := 0

	for _, col := range tableColumns {
		columnName := strings.ToLower(col.ColumnName)
		isPrimaryKey := strings.ToLower(strings.TrimSpace(col.Contype)) == primaryKeyKeyword
		if !isPrimaryKey && !slices.Contains(allowedColumns, columnName) {
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
		builder.WriteByte(' ')

		if strings.TrimSpace(col.ColumnNotNull) != "null" {
			builder.WriteString(col.ColumnNotNull)
			builder.WriteByte(' ')
		}

		if isPrimaryKey {
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
