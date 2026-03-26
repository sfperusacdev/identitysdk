package workflows_entities

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCheckTransition_TransitionNotFound(t *testing.T) {
	wf := Workflow{}

	req := TransitionRequest{
		Username: "user1",
		From:     "A",
		To:       "B",
	}

	err := wf.CheckTransition(req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheckTransition_NoPermissionsConfigured(t *testing.T) {
	wf := Workflow{
		WorkflowTransition: []WorkflowTransition{
			{Origen: "A", Destino: "B"},
		},
	}

	req := TransitionRequest{
		Username: "user1",
		From:     "A",
		To:       "B",
	}

	err := wf.CheckTransition(req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheckTransition_UserNotAuthorized(t *testing.T) {
	wf := Workflow{
		WorkflowTransition: []WorkflowTransition{
			{
				Origen:  "A",
				Destino: "B",
				WorkflowTransitionPermission: []WorkflowTransitionPermission{
					{UsuarioUsername: "user2"},
				},
			},
		},
	}

	req := TransitionRequest{
		Username: "user1",
		From:     "A",
		To:       "B",
	}

	err := wf.CheckTransition(req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheckTransition_SuccessWithoutLimits(t *testing.T) {
	wf := Workflow{
		WorkflowTransition: []WorkflowTransition{
			{
				Origen:  "A",
				Destino: "B",
				WorkflowTransitionPermission: []WorkflowTransitionPermission{
					{UsuarioUsername: "user1"},
				},
			},
		},
	}

	req := TransitionRequest{
		Username: "user1",
		From:     "A",
		To:       "B",
	}

	err := wf.CheckTransition(req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCheckTransition_DocumentAmountRequired(t *testing.T) {
	limit := decimal.NewFromInt(100)

	wf := Workflow{
		WorkflowTransition: []WorkflowTransition{
			{
				Origen:  "A",
				Destino: "B",
				WorkflowTransitionPermission: []WorkflowTransitionPermission{
					{
						UsuarioUsername:         "user1",
						LimiteMontoPorDocumento: &limit,
					},
				},
			},
		},
	}

	req := TransitionRequest{
		Username: "user1",
		From:     "A",
		To:       "B",
	}

	err := wf.CheckTransition(req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheckTransition_DocumentAmountExceeded(t *testing.T) {
	limit := decimal.NewFromInt(100)

	wf := Workflow{
		WorkflowTransition: []WorkflowTransition{
			{
				Origen:  "A",
				Destino: "B",
				WorkflowTransitionPermission: []WorkflowTransitionPermission{
					{
						UsuarioUsername:         "user1",
						LimiteMontoPorDocumento: &limit,
					},
				},
			},
		},
	}

	amount := decimal.NewFromInt(200)

	req := TransitionRequest{
		Username:       "user1",
		From:           "A",
		To:             "B",
		DocumentAmount: &amount,
	}

	err := wf.CheckTransition(req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheckTransition_PeriodAmountRequired(t *testing.T) {
	limit := decimal.NewFromInt(100)

	wf := Workflow{
		WorkflowTransition: []WorkflowTransition{
			{
				Origen:  "A",
				Destino: "B",
				WorkflowTransitionPermission: []WorkflowTransitionPermission{
					{
						UsuarioUsername:       "user1",
						LimiteMontoPorPeriodo: &limit,
					},
				},
			},
		},
	}

	req := TransitionRequest{
		Username: "user1",
		From:     "A",
		To:       "B",
	}

	err := wf.CheckTransition(req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheckTransition_PeriodAmountExceeded(t *testing.T) {
	limit := decimal.NewFromInt(100)

	wf := Workflow{
		WorkflowTransition: []WorkflowTransition{
			{
				Origen:  "A",
				Destino: "B",
				WorkflowTransitionPermission: []WorkflowTransitionPermission{
					{
						UsuarioUsername:       "user1",
						LimiteMontoPorPeriodo: &limit,
					},
				},
			},
		},
	}

	amount := decimal.NewFromInt(200)

	req := TransitionRequest{
		Username:     "user1",
		From:         "A",
		To:           "B",
		PeriodAmount: &amount,
	}

	err := wf.CheckTransition(req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheckTransition_SuccessWithAllLimits(t *testing.T) {
	docLimit := decimal.NewFromInt(100)
	periodLimit := decimal.NewFromInt(200)

	wf := Workflow{
		WorkflowTransition: []WorkflowTransition{
			{
				Origen:  "A",
				Destino: "B",
				WorkflowTransitionPermission: []WorkflowTransitionPermission{
					{
						UsuarioUsername:         "user1",
						LimiteMontoPorDocumento: &docLimit,
						LimiteMontoPorPeriodo:   &periodLimit,
					},
				},
			},
		},
	}

	docAmount := decimal.NewFromInt(50)
	periodAmount := decimal.NewFromInt(100)

	req := TransitionRequest{
		Username:       "user1",
		From:           "A",
		To:             "B",
		DocumentAmount: &docAmount,
		PeriodAmount:   &periodAmount,
	}

	err := wf.CheckTransition(req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
