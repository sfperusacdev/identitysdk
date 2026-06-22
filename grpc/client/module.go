package client

import (
	asistenciapb "github.com/sfperusacdev/identitysdk/grpc/gen/asistencia"
	contratospb "github.com/sfperusacdev/identitysdk/grpc/gen/contratos"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var Module = fx.Module("grpc-client",
	fx.Provide(
		fx.Annotate(
			NewGrpcClient,
			fx.As(new(grpc.ClientConnInterface)),
		),
		asistenciapb.NewTrabajadoresHorariosServiceClient,
		contratospb.NewAsistenciaMarcadorServiceClient,
		contratospb.NewTrabajadoresGRPCServiceClient,
	),
)
