package sark_services

import (
	"github.com/sfperusacdev/identitysdk/sark_services/asistencia"
	identityservice "github.com/sfperusacdev/identitysdk/sark_services/identity"
	"github.com/sfperusacdev/identitysdk/sark_services/storage"
	"github.com/sfperusacdev/identitysdk/sark_services/variables"
)

type SarkBridgeService struct {
	Identity   *identityservice.IdentityService
	Variables  *variables.VariablesService
	Storage    *storage.StorageService
	Asistencia *asistencia.AsistenciaService
}

func NewSarkBridgeService(
	Identity *identityservice.IdentityService,
	Variables *variables.VariablesService,
	Storage *storage.StorageService,
	Asistencia *asistencia.AsistenciaService,
) *SarkBridgeService {
	return &SarkBridgeService{
		Identity:   Identity,
		Variables:  Variables,
		Storage:    Storage,
		Asistencia: Asistencia,
	}
}
