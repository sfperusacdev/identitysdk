package workflows

import (
	"context"
	"fmt"
	"strings"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/helpers/workflows/workflows_entities"
	connection "github.com/sfperusacdev/identitysdk/pg-connection"
	"github.com/sfperusacdev/identitysdk/services"
	"github.com/shopspring/decimal"
	"github.com/user0608/goones/errs"
)

type DocumentWorkflowStateManager struct {
	manager connection.StorageManager
	bridge  *services.ExternalBridgeService
}

func NewDocumentWorkflowStateManager(
	manager connection.StorageManager,
	bridge *services.ExternalBridgeService,
) *DocumentWorkflowStateManager {
	return &DocumentWorkflowStateManager{
		manager: manager,
		bridge:  bridge,
	}
}

type DocumentState string

const (
	DocumentStatePending  DocumentState = "pending"
	DocumentStateApproved DocumentState = "approved"
	DocumentStateRejected DocumentState = "rejected"
)

type EntityConfig struct {
	TableName        string
	PrimaryKeyColumn string
	StateColumn      string
}

type StateChange struct {
	Doc            string
	TargetState    string
	DocumentAmount *decimal.Decimal
	PeriodAmount   *decimal.Decimal
}

type ChangeStateRequest struct {
	Username                string
	Entity                  EntityConfig
	Change                  StateChange
	Targets                 []string
	AllowPartial            bool
	RequireSameInitialState bool
	wf                      *workflows_entities.Workflow
	ifStateIsEmptyUse       DocumentState
}

type Entity struct {
	Code  string
	State string
}

func (m *DocumentWorkflowStateManager) ChangeState(ctx context.Context, req ChangeStateRequest) ([]string, error) {
	if len(req.Targets) == 0 {
		return []string{}, nil
	}

	if req.Username == "" {
		req.Username = identitysdk.Username(ctx)
	}
	if req.Entity.PrimaryKeyColumn == "" {
		req.Entity.PrimaryKeyColumn = "codigo"
	}
	if req.Entity.StateColumn == "" {
		req.Entity.StateColumn = "estado"
	}
	if req.wf == nil {
		wf, err := m.bridge.GetWorkflow(ctx, req.Change.Doc)
		if err != nil {
			return nil, errs.InternalErrorDirect(err.Error())
		}
		req.wf = wf
	}

	if req.wf.Codigo != req.Change.Doc {
		return nil, errs.BadRequestf(
			"el documento '%s' no coincide con el workflow '%s'",
			req.Change.Doc, req.wf.Codigo,
		)
	}

	var tx = m.manager.Conn(ctx)
	var records []Entity
	rs := tx.Table(req.Entity.TableName).Select(
		fmt.Sprintf("%s as code", req.Entity.PrimaryKeyColumn),
		fmt.Sprintf("%s as state", req.Entity.StateColumn),
	).Where(
		fmt.Sprintf("%s in ?", req.Entity.PrimaryKeyColumn),
		req.Targets,
	).Scan(&records)

	if rs.Error != nil {
		return nil, errs.Pgf(rs.Error)
	}

	var affectedTargets = []string{}
	var listErrors = []error{}
	var initialStates = make(map[string]struct{})
	for _, r := range records {
		var currentState = strings.TrimSpace(r.State)
		if currentState == "" {
			if req.ifStateIsEmptyUse != "" {
				currentState = string(req.ifStateIsEmptyUse)
			} else {
				currentState = string(DocumentStatePending)
			}
		}

		initialStates[currentState] = struct{}{}
		err := req.wf.CheckTransition(workflows_entities.TransitionRequest{
			Username:       req.Username,
			From:           identitysdk.Empresa(ctx, currentState),
			To:             identitysdk.Empresa(ctx, req.Change.TargetState),
			DocumentAmount: req.Change.DocumentAmount,
			PeriodAmount:   req.Change.PeriodAmount,
		})
		if err != nil {
			listErrors = append(listErrors, err)
			continue
		}
		affectedTargets = append(affectedTargets, r.Code)
	}
	if req.RequireSameInitialState && len(initialStates) > 1 {
		return nil, errs.BadRequestf("todos los documentos deben tener el mismo estado inicial")
	}
	if !req.AllowPartial && len(listErrors) > 0 {
		if len(req.Targets) == 1 {
			return nil, errs.BadRequestDirect(listErrors[0].Error())
		}
		return nil, errs.BadRequestf("no todos los documentos pueden ser actualizados")
	}
	if len(affectedTargets) == 0 {
		return []string{}, nil
	}
	rs = tx.Table(req.Entity.TableName).
		Where(fmt.Sprintf("%s in ?", req.Entity.PrimaryKeyColumn), affectedTargets).
		Update(req.Entity.StateColumn, identitysdk.RemovePrefix(req.Change.TargetState))
	if rs.Error != nil {
		return nil, errs.Pgf(rs.Error)
	}
	return affectedTargets, nil
}
