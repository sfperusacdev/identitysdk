package repos

import (
	"context"
	"database/sql"

	"github.com/sfperusacdev/identitysdk/setup/sqlsyncdata/descriptor"
	"github.com/user0608/goones/errs"
)

func (r *SQLTableRepository) GetTableColumns(ctx context.Context, tableName string) ([]descriptor.TableColumn, error) {
	if columns, ok := r.cache.Get(tableName); ok {
		return columns, nil
	}

	const query = `
	SELECT b.relname                                       as table_name,
            a.attname                                       as column_name,
            pg_catalog.format_type(a.atttypid, a.atttypmod) as column_type,
            CASE
                WHEN a.attnotnull THEN
                    'not null'
                ELSE
                    'null'
                END                                         as column_not_null,
            (select case when pc.contype = 'p' then 'primary key' when pc.contype = 'u' then 'unique' end
              from pg_catalog.pg_constraint pc
              where pc.contype in ('p', 'u')
                and pc.conrelid = attrelid
                and a.attnum = any (pc.conkey))              as contype
      FROM pg_catalog.pg_attribute a
              INNER JOIN
          (SELECT c.oid, c.relname
            FROM pg_catalog.pg_class c
                    LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
            WHERE c.relname = @table
              AND pg_catalog.pg_table_is_visible(c.oid)) b
          ON a.attrelid = b.oid
      WHERE a.attnum > 0
        AND NOT a.attisdropped;
	`

	tx := r.manager.Conn(ctx)
	var columns []descriptor.TableColumn
	rs := tx.Raw(query, sql.Named("table", tableName)).Scan(&columns)
	if rs.Error != nil {
		return nil, errs.Pgf(rs.Error)
	}
	r.cache.Set(tableName, columns)
	return columns, nil
}
