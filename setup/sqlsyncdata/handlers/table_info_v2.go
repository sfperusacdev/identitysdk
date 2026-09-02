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

type GetTableSqlInfoV2Handler struct {
	httpapi.MethodPost
	usecase  *usecase.SQLTableUsecase
	executor *domainexecutor.DomainExecutor
}

var _ httpapi.Route = (*GetTableSqlInfoV2Handler)(nil)

func NewGetTableSqlInfoV2Handler(lc fx.Lifecycle, usecase *usecase.SQLTableUsecase) *GetTableSqlInfoV2Handler {
	executor := domainexecutor.NewDefault()
	lc.Append(fx.Hook{OnStop: executor.Shutdown})
	return &GetTableSqlInfoV2Handler{usecase: usecase, executor: executor}
}

func (h *GetTableSqlInfoV2Handler) GetPath() string {
	return "/v2/sync_data/tabla_info"
}

func (h *GetTableSqlInfoV2Handler) HandleRequest(c echo.Context) error {
	var tables []string
	if err := binds.JSON(c, &tables); err != nil {
		return answer.Err(c, err)
	}
	ctx := c.Request().Context()
	domain := identitysdk.Empresa(ctx)
	var result []usecase.TableInfoV2Response
	if err := h.executor.Execute(ctx, domain, func(ctx context.Context) error {
		var err error
		result, err = h.usecase.GetTablesInfoV2(ctx, tables)
		return err
	}, nil); err != nil {
		return answer.Err(c, err)
	}
	return answer.Ok(c, result)
}
