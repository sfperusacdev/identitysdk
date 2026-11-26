package global

import (
	"github.com/sfperusacdev/identitysdk/sark_services/global/templates"
	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
)

type GlobalService struct {
	env       bridgeidentity.IdentityProvider
	Templates *templates.TemplatesService
}

func NewGlobalService(
	env bridgeidentity.IdentityProvider,
	templatesService *templates.TemplatesService,
) *GlobalService {
	return &GlobalService{
		env:       env,
		Templates: templatesService,
	}
}
