package sark_services

import (
	"testing"

	"github.com/sfperusacdev/identitysdk/configs"
	asistenciapb "github.com/sfperusacdev/identitysdk/grpc/gen/asistencia"
	"github.com/sfperusacdev/identitysdk/sark_services/global/templates"
	"github.com/sfperusacdev/identitysdk/sark_services/variables"
	"go.uber.org/fx"
)

func TestModuleProvidesGlobalTemplates(t *testing.T) {
	var bridge *SarkBridgeService
	var client *templates.TemplatesService
	app := fx.New(
		Module,
		fx.NopLogger,
		fx.Provide(
			func() configs.GeneralServiceConfigProvider { return nil },
			func() asistenciapb.TrabajadoresHorariosServiceClient { return nil },
		),
		fx.Replace(&variables.VariablesService{}),
		fx.Populate(&bridge, &client),
	)
	if err := app.Err(); err != nil {
		t.Fatal(err)
	}
	if bridge.Global == nil || bridge.Global.Templates != client {
		t.Fatal("bridge.Global.Templates is not the Fx-provided client")
	}
}
