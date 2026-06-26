package sark_services

import (
	"github.com/sfperusacdev/identitysdk/sark_services/asistencia"
	"github.com/sfperusacdev/identitysdk/sark_services/storage"
	"github.com/sfperusacdev/identitysdk/sark_services/variables"
)

type SarkBridgeService struct {
	Variables  *variables.VariablesService
	Storage    *storage.StorageService
	Asistencia *asistencia.AsistenciaService
}

func NewSarkBridgeService(
	Variables *variables.VariablesService,
	Storage *storage.StorageService,
	Asistencia *asistencia.AsistenciaService,
) *SarkBridgeService {
	return &SarkBridgeService{
		Variables:  Variables,
		Storage:    Storage,
		Asistencia: Asistencia,
	}
}
