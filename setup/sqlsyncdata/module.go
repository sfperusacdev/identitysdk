package sqlsyncdata

import (
	"github.com/sfperusacdev/identitysdk/httpapi"
	"github.com/sfperusacdev/identitysdk/setup/sqlsyncdata/descriptor"
	"github.com/sfperusacdev/identitysdk/setup/sqlsyncdata/handlers"
	"github.com/sfperusacdev/identitysdk/setup/sqlsyncdata/repos"
	"github.com/sfperusacdev/identitysdk/setup/sqlsyncdata/usecase"
	"go.uber.org/fx"
)

func LoadModule(descriptors ...descriptor.TableDescriptor) fx.Option {
	return fx.Module("sqlsyncdata",
		fx.Supply(usecase.TableDescriptors(descriptors)),
		fx.Provide(repos.NewSQLTableRepository),
		fx.Provide(usecase.NewSQLTableUsecase),
		fx.Provide(
			httpapi.AsRoute(handlers.NewGetTableSqlInfoHandler),
			httpapi.AsRoute(handlers.NewSqlTableSyncDataHandler),
		),
	)
}
