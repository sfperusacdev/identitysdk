package repos

import (
	"github.com/sfperusacdev/identitysdk/helpers/staticstore"
	connection "github.com/sfperusacdev/identitysdk/pg-connection"
	"github.com/sfperusacdev/identitysdk/setup/sqlsyncdata/descriptor"
)

type SQLTableRepository struct {
	manager connection.StorageManager
	cache   *staticstore.StaticStore[string, []descriptor.TableColumn]
}

func NewSQLTableRepository(manager connection.StorageManager) *SQLTableRepository {
	return &SQLTableRepository{
		manager: manager,
		cache:   staticstore.New[string, []descriptor.TableColumn](),
	}
}
