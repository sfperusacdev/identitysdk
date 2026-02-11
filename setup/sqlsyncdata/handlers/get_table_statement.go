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
)

type GetTableSqlInfoHandler struct {
	httpapi.MethodPost
	usecase  *usecase.SQLTableUsecase
	executor *domainexecutor.DomainExecutor
}

var _ httpapi.Route = (*GetTableSqlInfoHandler)(nil)

func NewGetTableSqlInfoHandler(usecase *usecase.SQLTableUsecase) *GetTableSqlInfoHandler {
	return &GetTableSqlInfoHandler{
		usecase:  usecase,
		executor: domainexecutor.NewDefault(),
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
