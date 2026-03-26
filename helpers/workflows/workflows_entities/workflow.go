package workflows_entities

import (
	"errors"

	"github.com/shopspring/decimal"
)

type Workflow struct {
	Codigo             string               `json:"codigo"`
	Descripcion        string               `json:"descripcion"`
	Color              string               `json:"color"`
	WorkflowState      []WorkflowState      `json:"workflow_state"`
	WorkflowTransition []WorkflowTransition `json:"workflow_transition"`
}

type WorkflowState struct {
	Codigo      string  `json:"codigo"`
	Nombre      string  `json:"nombre"`
	Tipo        string  `json:"tipo"`
	Color       *string `json:"color"`
	EsInmutable bool    `json:"es_inmutable"`
}

type WorkflowTransition struct {
	Codigo                       string                         `json:"codigo"`
	Origen                       string                         `json:"origen"`
	Destino                      string                         `json:"destino"`
	WorkflowTransitionPermission []WorkflowTransitionPermission `json:"workflow_transition_permission"`
}

type WorkflowTransitionPermission struct {
	UsuarioUsername         string           `json:"usuario_username"`
	LimiteMontoPorDocumento *decimal.Decimal `json:"limite_monto_por_documento"`
	LimitePeriodoTipo       *string          `json:"limite_periodo_tipo"`
	LimiteMontoPorPeriodo   *decimal.Decimal `json:"limite_monto_por_periodo"`
}
type TransitionRequest struct {
	Username       string
	From           string
	To             string
	DocumentAmount *decimal.Decimal
	PeriodAmount   *decimal.Decimal
}

func (wf Workflow) getState(codigo string) *WorkflowState {
	for i := range wf.WorkflowState {
		if wf.WorkflowState[i].Codigo == codigo {
			return &wf.WorkflowState[i]
		}
	}
	return nil
}

func (wf Workflow) CheckTransition(req TransitionRequest) error {
	fromState := wf.getState(req.From)
	if fromState == nil {
		return errors.New("estado origen no existe")
	}

	if fromState.EsInmutable {
		return errors.New("no se puede realizar la transición desde un estado inmutable")
	}

	var transition *WorkflowTransition

	for i := range wf.WorkflowTransition {
		t := &wf.WorkflowTransition[i]
		if t.Origen == req.From && t.Destino == req.To {
			transition = t
			break
		}
	}

	if transition == nil {
		return errors.New("no existe una transición válida desde el estado origen al estado destino")
	}

	if len(transition.WorkflowTransitionPermission) == 0 {
		return errors.New("la transición no tiene permisos configurados")
	}

	for _, p := range transition.WorkflowTransitionPermission {

		if p.UsuarioUsername != req.Username {
			continue
		}

		if err := p.Validate(req); err != nil {
			return err
		}

		return nil
	}

	return errors.New("el usuario no está autorizado para ejecutar esta transición")
}

func (p WorkflowTransitionPermission) Validate(req TransitionRequest) error {
	if p.LimiteMontoPorDocumento != nil {
		if req.DocumentAmount == nil {
			return errors.New("se requiere el monto del documento para validar la transición")
		}
		if req.DocumentAmount.GreaterThan(*p.LimiteMontoPorDocumento) {
			return errors.New("el monto del documento excede el límite permitido para esta transición")
		}
	}

	if p.LimiteMontoPorPeriodo != nil {
		if req.PeriodAmount == nil {
			return errors.New("se requiere el monto acumulado del periodo para validar la transición")
		}
		if req.PeriodAmount.GreaterThan(*p.LimiteMontoPorPeriodo) {
			return errors.New("el monto acumulado del periodo excede el límite permitido para esta transición")
		}
	}

	return nil
}
