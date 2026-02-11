package handlers

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/binds"
	"github.com/sfperusacdev/identitysdk/helpers/domainexecutor"
	"github.com/sfperusacdev/identitysdk/httpapi"
	"github.com/sfperusacdev/identitysdk/setup/sqlsyncdata/usecase"
	"github.com/user0608/goones/answer"
	"go.uber.org/fx"
)

type SqlTableSyncDataHandler struct {
	httpapi.MethodPost
	usecase  *usecase.SQLTableUsecase
	executor *domainexecutor.DomainExecutor
}

var _ httpapi.Route = (*SqlTableSyncDataHandler)(nil)

func NewSqlTableSyncDataHandler(lc fx.Lifecycle, usecase *usecase.SQLTableUsecase) *SqlTableSyncDataHandler {
	executor := domainexecutor.NewDefault()
	lc.Append(fx.Hook{OnStop: executor.Shutdown})
	return &SqlTableSyncDataHandler{
		usecase:  usecase,
		executor: executor,
	}
}

func (h *SqlTableSyncDataHandler) GetPath() string {
	return "/v1/sync_data/sync"
}

func (h *SqlTableSyncDataHandler) HandleRequest(c echo.Context) error {
	var ctx = c.Request().Context()
	var domain = identitysdk.Empresa(ctx)

	var syncRequest usecase.TableSyncRequest
	if err := binds.JSON(c, &syncRequest); err != nil {
		return answer.Err(c, err)
	}

	var result *usecase.TableSyncResponse
	var executeDomain = fmt.Sprintf("%s.%s", domain, syncRequest.TableName)
	if err := h.executor.Execute(ctx, executeDomain, func(ctx context.Context) error {
		var err error
		result, err = h.usecase.SyncTable(ctx, domain, syncRequest)
		return err
	}); err != nil {
		return answer.Err(c, err)
	}
	return answer.Ok(c, result)
}
