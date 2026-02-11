package repos

import (
	"context"
)

func (r *SQLTableRepository) GetTablePrimaryKeys(ctx context.Context, tableName string) ([]string, error) {
	var primarykeys []string
	columns, err := r.GetTableColumns(ctx, tableName)
	if err != nil {
		return nil, err
	}
	for _, c := range columns {
		if c.IsPrimaryKey() {
			primarykeys = append(primarykeys, c.ColumnName)
		}
	}
	return primarykeys, nil
}
