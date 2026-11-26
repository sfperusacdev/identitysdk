package global

import (
	"testing"

	"github.com/sfperusacdev/identitysdk/sark_services/global/templates"
	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
)

func TestNewGlobalService(t *testing.T) {
	env := bridgeidentity.NewIdentityProvider()
	client := templates.NewTemplatesService(env)
	if got := NewGlobalService(env, client).Templates; got != client {
		t.Fatalf("GlobalService.Templates = %p, want %p", got, client)
	}
}
