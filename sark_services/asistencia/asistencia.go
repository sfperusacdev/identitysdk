package asistencia

import asistenciapb "github.com/sfperusacdev/identitysdk/grpc/gen/asistencia"

type AsistenciaService struct {
	asistenciaGrpc asistenciapb.TrabajadoresHorariosServiceClient
}

func NewAsistenciaService(
	asistenciaGrpc asistenciapb.TrabajadoresHorariosServiceClient,
) *AsistenciaService {
	return &AsistenciaService{
		asistenciaGrpc: asistenciaGrpc,
	}
}
