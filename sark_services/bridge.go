package sark_services

import (
	"github.com/sfperusacdev/identitysdk/sark_services/asistencia"
	"github.com/sfperusacdev/identitysdk/sark_services/global"
	identityservice "github.com/sfperusacdev/identitysdk/sark_services/identity"
	"github.com/sfperusacdev/identitysdk/sark_services/storage"
	"github.com/sfperusacdev/identitysdk/sark_services/variables"
)

type SarkBridgeService struct {
	Env        identityservice.IdentityProvider
	Identity   *identityservice.IdentityService
	Variables  *variables.VariablesService
	Storage    *storage.StorageService
	Asistencia *asistencia.AsistenciaService
	Global     *global.GlobalService
}

func NewSarkBridgeService(
	Env identityservice.IdentityProvider,
	Identity *identityservice.IdentityService,
	Variables *variables.VariablesService,
	Storage *storage.StorageService,
	Asistencia *asistencia.AsistenciaService,
	Global *global.GlobalService,
) *SarkBridgeService {
	return &SarkBridgeService{
		Env:        Env,
		Identity:   Identity,
		Variables:  Variables,
		Storage:    Storage,
		Asistencia: Asistencia,
		Global:     Global,
	}
}
