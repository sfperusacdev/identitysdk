package global

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sfperusacdev/identitysdk/helpers/workflows/workflows_entities"
	"github.com/sfperusacdev/identitysdk/xreq"
)

func (s *GlobalService) GetWorkflow(ctx context.Context, doc string) (*workflows_entities.Workflow, error) {
	baseURL, err := s.env.GetGeneralServiceURL(ctx, s.env.Empresa(ctx))
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
		xreq.WithAuthorization(s.env.SessionToken(ctx)),
		xreq.WithUnmarshalResponseInto(&response),
		xreq.WithJsonContentType(),
	); err != nil {
		return nil, err
	}

	return &response.Data, nil
}
