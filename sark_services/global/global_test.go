package global

import (
	"testing"

	bridgeidentity "github.com/sfperusacdev/identitysdk/sark_services/identity"
)

func TestNewGlobalServiceIncludesTemplates(t *testing.T) {
	if NewGlobalService(bridgeidentity.NewIdentityProvider()).Templates == nil {
		t.Fatal("GlobalService.Templates is nil")
	}
}
