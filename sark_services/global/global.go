package global

import (
	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
)

type GlobalService struct {
	env bridgeidentity.IdentityProvider
}

func NewGlobalService(env bridgeidentity.IdentityProvider) *GlobalService {
	return &GlobalService{env: env}
}
