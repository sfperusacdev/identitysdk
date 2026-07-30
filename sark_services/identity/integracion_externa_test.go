package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitIntegrationURL(t *testing.T) {
	tests := map[string]struct {
		value    string
		readOnly bool
		outValue string
	}{
		"simple-01": {
			value:    "http://localhost:808",
			readOnly: false,
			outValue: "http://localhost:808",
		},
		"simple-readonly": {
			value:    "http://localhost:8080:ro",
			readOnly: true,
			outValue: "http://localhost:8080",
		},
		"simple-readonly2": {
			value:    "http://sfperusac.com:ro",
			readOnly: true,
			outValue: "http://sfperusac.com",
		},
		"readonly2": {
			value:    "http://sfperusac.com/:ro",
			readOnly: true,
			outValue: "http://sfperusac.com",
		},
		"readonly4": {
			value:    "http://sfperusac.com:8080/:ro",
			readOnly: true,
			outValue: "http://sfperusac.com:8080",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			outValue, readOnly := splitIntegrationURL(test.value)
			assert.Equal(t, test.readOnly, readOnly)
			assert.Equal(t, test.outValue, outValue)
		})
	}
}
