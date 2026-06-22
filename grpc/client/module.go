package client

import (
	"github.com/sfperusacdev/identitysdk/configs"
	asistenciapb "github.com/sfperusacdev/identitysdk/grpc/gen/asistencia"
	contratospb "github.com/sfperusacdev/identitysdk/grpc/gen/contratos"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func newGrpcClient(resourceCode string) any {
	return func(
		lc fx.Lifecycle,
		config configs.GeneralServiceConfigProvider,
	) grpc.ClientConnInterface {
		return NewGrpcClient(lc, resourceCode, config)
	}
}

var contratosModule = fx.Module("grpc-client-contratos",
	fx.Provide(
		newGrpcClient("com.sfperusac.contratos"),
		fx.Private,
	),
	fx.Provide(
		asistenciapb.NewTrabajadoresHorariosServiceClient,
		contratospb.NewAsistenciaMarcadorServiceClient,
		contratospb.NewTrabajadoresGRPCServiceClient,
	),
)

var Module = fx.Module(
	"grpc-client",
	contratosModule,
)
