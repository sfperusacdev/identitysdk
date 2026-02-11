package handlers

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/binds"
	"github.com/sfperusacdev/identitysdk/helpers/domainexecutor"
	"github.com/sfperusacdev/identitysdk/httpapi"
	"github.com/sfperusacdev/identitysdk/setup/sqlsyncdata/usecase"
	"github.com/user0608/goones/answer"
	"go.uber.org/fx"
)

type GetTableSqlInfoHandler struct {
	httpapi.MethodPost
	usecase  *usecase.SQLTableUsecase
	executor *domainexecutor.DomainExecutor
}

var _ httpapi.Route = (*GetTableSqlInfoHandler)(nil)

func NewGetTableSqlInfoHandler(lc fx.Lifecycle, usecase *usecase.SQLTableUsecase) *GetTableSqlInfoHandler {
	executor := domainexecutor.NewDefault()
	lc.Append(fx.Hook{OnStop: executor.Shutdown})

	return &GetTableSqlInfoHandler{
		usecase:  usecase,
		executor: executor,
	}
}

func (h *GetTableSqlInfoHandler) GetPath() string {
	return "/v1/sync_data/tabla_info"
}

func (h *GetTableSqlInfoHandler) HandleRequest(c echo.Context) error {
	var tables []string
	if err := binds.JSON(c, &tables); err != nil {
		return answer.Err(c, err)
	}
	var ctx = c.Request().Context()
	var domain = identitysdk.Empresa(ctx)
	var result []usecase.TableInfoResponse
	if err := h.executor.Execute(ctx, domain, func(ctx context.Context) error {
		var err error
		result, err = h.usecase.GetTablesStatement(ctx, tables)
		return err
	}); err != nil {
		return answer.Err(c, err)
	}
	return answer.Ok(c, result)
}
