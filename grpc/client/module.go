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
		contratospb.NewAsistenciaMarcadorServiceClient,
		contratospb.NewTrabajadoresGRPCServiceClient,
	),
)

var asistenciaModule = fx.Module("grpc-client-asistencia",
	fx.Provide(
		newGrpcClient("com.sfperusac.asistencia"),
		fx.Private,
	),
	fx.Provide(
		asistenciapb.NewTrabajadoresHorariosServiceClient,
	),
)

var Module = fx.Module(
	"grpc-client",
	contratosModule,
	asistenciaModule,
)
