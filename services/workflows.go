package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/helpers/workflows/workflows_entities"
	"github.com/sfperusacdev/identitysdk/xreq"
)

func (s *ExternalBridgeService) GetWorkflow(ctx context.Context, doc string) (*workflows_entities.Workflow, error) {
	company, token := s.readCompanyAndToken(ctx)
	baseURL, err := identitysdk.GetGeneralServiceURL(ctx, company)
	if err != nil {
		slog.Error("error retrieving service URL", "error", err)
		return nil, err
	}

	var response struct {
		Message string                      `json:"message"`
		Data    workflows_entities.Workflow `json:"data"`
	}

	endpoint := fmt.Sprintf("/api/v2/documentos/workflows/%s", doc)

	if err := xreq.MakeRequest(ctx,
		baseURL, endpoint,
		xreq.WithAuthorization(token),
		xreq.WithUnmarshalResponseInto(&response),
		xreq.WithJsonContentType(),
	); err != nil {
		return nil, err
	}

	return &response.Data, nil
}
