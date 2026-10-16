package asistencia

import (
	asistenciapb "github.com/sfperusacdev/identitysdk/grpc/gen/asistencia"
	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
)

type AsistenciaService struct {
	asistenciaGrpc asistenciapb.TrabajadoresHorariosServiceClient
	env            bridgeidentity.IdentityProvider
}

func NewAsistenciaService(
	asistenciaGrpc asistenciapb.TrabajadoresHorariosServiceClient,
) *AsistenciaService {
	return &AsistenciaService{
		asistenciaGrpc: asistenciaGrpc,
		env:            bridgeidentity.NewIdentityProvider(),
	}
}
