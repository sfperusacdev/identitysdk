package mmsql

import (
	"fmt"
	"io/fs"

	"github.com/sfperusacdev/identitysdk/utils/sql/sqlproc"
	"github.com/sfperusacdev/identitysdk/utils/sql/sqlreader"
	"github.com/sfperusacdev/identitysdk/utils/sql/sqlutil"
)

type StoredProcedureStore struct {
	procedures map[string]string
}

func NewStoredProcedureStore(filesystem fs.FS) func() (*StoredProcedureStore, error) {
	return func() (*StoredProcedureStore, error) {
		files, err := sqlreader.LoadSQLFiles(filesystem, ".")
		if err != nil {
			return nil, err
		}

		procedures := make(map[string]string, len(files))

		for _, file := range files {
			if err := sqlproc.ValidateProcedureDefinition(file.Content); err != nil {
				return nil, fmt.Errorf("%s: %w", file.Path, err)
			}

			procedureName, err := sqlproc.ExtractProcedureName(file.Content)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", file.Path, err)
			}

			normalizedName, err := sqlutil.NormalizeSQLServerIdentifier(procedureName)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", file.Path, err)
			}

			procedures[normalizedName] = file.Content
		}

		return &StoredProcedureStore{
			procedures: procedures,
		}, nil
	}
}

func (s *StoredProcedureStore) Get(name string) (string, bool) {
	source, ok := s.procedures[name]
	return source, ok
}
