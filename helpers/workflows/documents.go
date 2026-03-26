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
	Username          string
	Entity            EntityConfig
	Change            StateChange
	Targets           []string
	AllowPartial      bool
	wf                *workflows_entities.Workflow
	ifStateIsEmptyUse DocumentState
}

type Entity struct {
	Code  string
	State string
}

func (m *DocumentWorkflowStateManager) ChangeState(ctx context.Context, req ChangeStateRequest) error {
	if len(req.Targets) == 0 {
		return nil
	}

	if req.Username == "" {
		req.Username = identitysdk.Username(ctx)
	}
	if req.Entity.PrimaryKeyColumn == "" {
		req.Entity.PrimaryKeyColumn = "codigo"
	}
	if req.Entity.StateColumn == "" {
		req.Entity.PrimaryKeyColumn = "estado"
	}
	if req.wf == nil {
		wf, err := m.bridge.GetWorkflow(ctx, req.Change.Doc)
		if err != nil {
			return errs.InternalErrorDirect(err.Error())
		}
		req.wf = wf
	}

	if req.wf.Codigo != req.Change.Doc {
		return errs.BadRequestf(
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
		return errs.Pgf(rs.Error)
	}

	var affectedTargets = []string{}
	var listErrors = []error{}

	for _, r := range records {
		var currentState = strings.TrimSpace(r.State)
		if currentState == "" {
			currentState = string(DocumentStatePending)
		}
		err := req.wf.CheckTransition(workflows_entities.TransitionRequest{
			Username:       req.Username,
			From:           identitysdk.Empresa(ctx, r.State),
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

	if !req.AllowPartial && len(listErrors) > 0 {
		if len(req.Targets) == 1 {
			return errs.BadRequestDirect(listErrors[0].Error())
		}
		return errs.BadRequestf("no todos los documentos pueden ser actualizados")
	}

	if len(affectedTargets) > 0 {
		rs := tx.Table(req.Entity.TableName).
			Where(fmt.Sprintf("%s in ?", req.Entity.PrimaryKeyColumn), affectedTargets).
			Update(req.Entity.StateColumn, identitysdk.RemovePrefix(req.Change.TargetState))

		return errs.Pgf(rs.Error)
	}

	return nil
}
