package repos

import (
	"context"

	"github.com/sfperusacdev/identitysdk/setup/sqlsyncdata/descriptor"
	"github.com/user0608/goones/errs"
	"gorm.io/gorm/clause"
)

func (r *SQLTableRepository) GetTableData(ctx context.Context, domain string, desc descriptor.TableDescriptor, syncAt int64) ([]map[string]any, error) {
	columns, err := r.GetTableColumns(ctx, desc.Table)
	if err != nil {
		return nil, err
	}
	query, args := desc.BuildSelectStatement(columns, domain, syncAt)
	var tx = r.manager.Conn(ctx)
	var rows = []map[string]any{}
	rs := tx.Raw(query, args...).Scan(&rows)
	if rs.Error != nil {
		return nil, errs.Pgf(rs.Error)
	}
	return rows, nil
}

func (r *SQLTableRepository) InsertData(ctx context.Context, tableName string, rows []map[string]any) error {
	if len(rows) == 0 {
		return nil
	}

	db := r.manager.Conn(ctx)

	pkColumns, err := r.GetTablePrimaryKeys(ctx, tableName)
	if err != nil {
		return err
	}

	conflictColumns := make([]clause.Column, len(pkColumns))
	for i := range pkColumns {
		conflictColumns[i] = clause.Column{Name: pkColumns[i]}
	}

	updateColumns := make([]string, 0, len(rows[0]))
	for key := range rows[0] {
		updateColumns = append(updateColumns, key)
	}

	result := db.Clauses(
		clause.OnConflict{
			UpdateAll: true,
			Columns:   conflictColumns,
			DoUpdates: clause.AssignmentColumns(updateColumns),
		},
	).Table(tableName).CreateInBatches(rows, 100)

	if result.Error != nil {
		return errs.Pgf(result.Error)
	}

	return nil

}
