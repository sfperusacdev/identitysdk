package identitysdk_test

import (
	"context"
	"testing"

	"github.com/sfperusacdev/identitysdk"
)

func TestEmpresa(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		suffixes []string
		expected string
	}{
		{
			name:     "NoSuffix",
			ctx:      createContext(),
			suffixes: nil,
			expected: "sfperu",
		},
		{
			name:     "EmptySingleSuffix",
			ctx:      createContext(),
			suffixes: []string{""},
			expected: "sfperu",
		},
		{
			name:     "Prefix",
			ctx:      createContext(),
			suffixes: []string{"%"},
			expected: "sfperu.%",
		},
		{
			name:     "WithSuffixes",
			ctx:      createContext(),
			suffixes: []string{"c1", "c2", "c3"},
			expected: "sfperu.c1.c2.c3",
		},
		{
			name:     "FromCreateContextImplícito",
			ctx:      createContext(),
			suffixes: []string{"c1"},
			expected: "sfperu.c1",
		},
		{
			name:     "MissingDomainKey",
			ctx:      context.Background(),
			suffixes: []string{"c1"},
			expected: "####empresa-no-found####.c1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := identitysdk.Empresa(tt.ctx, tt.suffixes...)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSucursal(t *testing.T) {
	type key string

	tests := []struct {
		name     string
		ctx      context.Context
		suffixes []string
		expected string
	}{
		{
			name:     "NoSuffix",
			ctx:      createContext(),
			suffixes: nil,
			expected: "",
		},
		{
			name:     "GenLikePrefix",
			ctx:      createContext(),
			suffixes: []string{"%"},
			expected: "sfperu.sf001.%",
		},
		{
			name:     "EmptySingleSuffix",
			ctx:      createContext(),
			suffixes: []string{""},
			expected: "",
		},
		{
			name:     "WithSuffixes",
			ctx:      createContext(),
			suffixes: []string{"c1", "c2"},
			expected: "sfperu.sf001.c1.c2",
		},
		{
			name:     "MissingSucursalKey",
			ctx:      context.WithValue(context.Background(), key("key"), "sfperu"),
			suffixes: []string{"c1"},
			expected: "####empresa-no-found####.####sucursal-no-found####.c1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := identitysdk.Sucursal(tt.ctx, tt.suffixes...)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
