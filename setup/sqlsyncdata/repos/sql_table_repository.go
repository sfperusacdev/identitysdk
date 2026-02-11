package repos

import (
	connection "github.com/sfperusacdev/identitysdk/pg-connection"
)

type SQLTableRepository struct {
	manager connection.StorageManager
}

func NewSQLTableRepository(manager connection.StorageManager) *SQLTableRepository {
	return &SQLTableRepository{manager: manager}
}
