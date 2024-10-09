package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_integracionExternaURlSplit(t *testing.T) {
	var bridge = ExternalBridgeService{}
	var dts = map[string]struct {
		value    string
		readOnly bool
		outValue string
	}{
		"simple-01": {
			value:    "http://localhost:808",
			readOnly: false,
			outValue: "http://localhost:808",
		},
		"simple-readony": {
			value:    "http://localhost:8080:ro",
			readOnly: true,
			outValue: "http://localhost:8080",
		},
		"simple-readony2": {
			value:    "http://sfperusac.com:ro",
			readOnly: true,
			outValue: "http://sfperusac.com",
		},
		"readony2": {
			value:    "http://sfperusac.com/:ro",
			readOnly: true,
			outValue: "http://sfperusac.com",
		},
		"readony4": {
			value:    "http://sfperusac.com:8080/:ro",
			readOnly: true,
			outValue: "http://sfperusac.com:8080",
		},
	}

	for name, dt := range dts {
		t.Run(name, func(t *testing.T) {
			outvalue, readOnly := bridge.integracionExternaURlSplit(dt.value)
			assert.Equal(t, dt.readOnly, readOnly)
			assert.Equal(t, dt.outValue, outvalue)
		})
	}
}
