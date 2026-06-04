package mmsql

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/sfperusacdev/identitysdk/utils/sql/sqlproc"
	"github.com/sfperusacdev/identitysdk/utils/sql/sqlutil"
	"github.com/user0608/goones/errs"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type StoredProcedureExecutor struct {
	connection *SQLServerConnection
	store      *StoredProcedureStore
}

func NewStoredProcedureExecutor(
	connection *SQLServerConnection,
	store *StoredProcedureStore,
) *StoredProcedureExecutor {
	return &StoredProcedureExecutor{
		connection: connection,
		store:      store,
	}
}

type StoredProcedureResult struct {
	rs                      *gorm.DB
	tx                      *gorm.DB
	temporaryProcedureName  string
	normalizedProcedureName string
	dropped                 bool
}

func (r *StoredProcedureResult) dropTemporaryProcedure() {
	if r.dropped {
		return
	}

	r.dropped = true

	dropResult := r.tx.Session(
		&gorm.Session{Logger: logger.Default.LogMode(logger.Error)},
	).Exec(fmt.Sprintf("DROP PROCEDURE %s", r.temporaryProcedureName))
	if dropResult.Error != nil {
		slog.Error(
			"failed to drop temporary stored procedure",
			"procedure_name", r.normalizedProcedureName,
			"temporary_procedure_name", r.temporaryProcedureName,
			"error", dropResult.Error,
		)
	}
}

func (r *StoredProcedureResult) Rows() (*sql.Rows, error) {
	return r.rs.Rows()
}

func (r *StoredProcedureResult) Row() *sql.Row {
	defer r.dropTemporaryProcedure()
	return r.rs.Row()
}

func (r *StoredProcedureResult) Scan(dest any) *gorm.DB {
	defer r.dropTemporaryProcedure()
	return r.rs.Scan(dest)
}

func (r *StoredProcedureResult) ScanRows(rows *sql.Rows, dest any) error {
	return r.rs.ScanRows(rows, dest)
}

func (r *StoredProcedureResult) Error() error {
	return r.rs.Error
}

func (r *StoredProcedureResult) Close() {
	r.dropTemporaryProcedure()
}

func (e *StoredProcedureExecutor) Execute(
	ctx context.Context,
	query string,
	values ...any,
) (*StoredProcedureResult, error) {
	tx := e.connection.Conn(ctx)

	procName, err := sqlproc.GetStoredProcedureIdentifierFromQuery(query)
	if err != nil {
		slog.Error(
			"failed to get stored procedure identifier",
			"query", query,
			"error", err,
		)
		return nil, errs.InternalErrorDirect(errs.ErrInternal)
	}

	normalizedName, err := sqlutil.NormalizeSQLServerIdentifier(procName)
	if err != nil {
		slog.Error(
			"failed to normalize stored procedure identifier",
			"procedure_name", procName,
			"error", err,
		)
		return nil, errs.InternalErrorDirect(errs.ErrInternal)
	}

	statement, ok := e.store.Get(normalizedName)
	if !ok {
		slog.Error(
			"stored procedure not found",
			"procedure_name", normalizedName,
		)
		return nil, errs.InternalErrorDirect(errs.ErrInternal)
	}

	definition, err := sqlproc.RenameProcedureWithRandomName(statement)
	if err != nil {
		slog.Error(
			"failed to generate temporary stored procedure definition",
			"procedure_name", normalizedName,
			"error", err,
		)
		return nil, errs.InternalErrorDirect(errs.ErrInternal)
	}

	createRes := tx.Session(
		&gorm.Session{Logger: logger.Default.LogMode(logger.Error)},
	).Exec(definition.SqlDefinition)
	if createRes.Error != nil {
		slog.Error(
			"failed to create temporary stored procedure",
			"procedure_name", normalizedName,
			"temporary_procedure_name", definition.Name,
			"sql_definition", definition.SqlDefinition,
			"error", createRes.Error,
		)
		return nil, errs.InternalErrorDirect(errs.ErrInternal)
	}

	shouldDrop := true
	defer func() {
		if !shouldDrop {
			return
		}

		dropRes := tx.Session(
			&gorm.Session{Logger: logger.Default.LogMode(logger.Error)},
		).Exec(fmt.Sprintf("DROP PROCEDURE %s", definition.Name))
		if dropRes.Error != nil {
			slog.Error(
				"failed to drop temporary stored procedure",
				"procedure_name", normalizedName,
				"temporary_procedure_name", definition.Name,
				"error", dropRes.Error,
			)
		}
	}()

	query, err = sqlproc.ReplaceStoredProcedureIdentifierInQuery(query, definition.Name)
	if err != nil {
		slog.Error(
			"failed to replace stored procedure identifier in query",
			"procedure_name", normalizedName,
			"temporary_procedure_name", definition.Name,
			"query", query,
			"error", err,
		)
		return nil, errs.InternalErrorDirect(errs.ErrInternal)
	}

	res := tx.Raw(query, values...)
	if res.Error != nil {
		slog.Error(
			"failed to execute stored procedure query",
			"procedure_name", normalizedName,
			"temporary_procedure_name", definition.Name,
			"query", query,
			"values", values,
			"error", res.Error,
		)
		return nil, errs.InternalError(res.Error, "no se pudo ejecutar la consulta")
	}

	shouldDrop = false

	return &StoredProcedureResult{
		tx:                      tx,
		rs:                      res,
		temporaryProcedureName:  definition.Name,
		normalizedProcedureName: normalizedName,
	}, nil
}
