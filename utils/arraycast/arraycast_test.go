package arraycast

import (
	"testing"
	"time"
)

func TestGetStringAt(t *testing.T) {
	arr := []string{" uno ", "  ", "tres"}
	tests := []struct {
		name     string
		index    int
		expected string
	}{
		{"valor válido con espacios", 0, "uno"},
		{"valor vacío", 1, ""},
		{"índice fuera de rango positivo", 3, ""},
		{"índice fuera de rango negativo", -1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStringAt(arr, tt.index)
			if got != tt.expected {
				t.Errorf("esperado '%s', obtenido '%s'", tt.expected, got)
			}
		})
	}
}

func TestGetIntAt(t *testing.T) {
	arr := []string{"10", "  -5 ", "abc", "", "   "}
	tests := []struct {
		name     string
		index    int
		expected int
	}{
		{"entero válido", 0, 10},
		{"entero negativo válido", 1, -5},
		{"cadena inválida", 2, 0},
		{"vacío", 3, 0},
		{"espacios", 4, 0},
		{"fuera de rango", 6, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetIntAt(arr, tt.index)
			if got != tt.expected {
				t.Errorf("esperado %d, obtenido %d", tt.expected, got)
			}
		})
	}
}

func TestGetFloatAt(t *testing.T) {
	arr := []string{"3.14", " -2.5 ", "abc", "", " "}
	tests := []struct {
		name     string
		index    int
		expected float64
	}{
		{"float válido", 0, 3.14},
		{"float negativo válido", 1, -2.5},
		{"texto inválido", 2, 0},
		{"vacío", 3, 0},
		{"espacios", 4, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFloatAt(arr, tt.index)
			if got != tt.expected {
				t.Errorf("esperado %f, obtenido %f", tt.expected, got)
			}
		})
	}
}

func TestGetBoolAt(t *testing.T) {
	arr := []string{"true", "False", "yes", "no", "1", "0", "", "   "}
	tests := []struct {
		name     string
		index    int
		expected bool
	}{
		{"true literal", 0, true},
		{"false literal", 1, false},
		{"yes", 2, true},
		{"no", 3, false},
		{"1", 4, true},
		{"0", 5, false},
		{"vacío", 6, false},
		{"espacios", 7, false},
		{"fuera de rango", 9, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBoolAt(arr, tt.index)
			if got != tt.expected {
				t.Errorf("esperado %t, obtenido %t", tt.expected, got)
			}
		})
	}
}

func TestGetDateAt(t *testing.T) {
	now := time.Now()
	dateStr := now.Format("2006-01-02 15:04:05")
	arr := []string{dateStr, "2020-12-31", "invalid", "", "   "}

	tests := []struct {
		name     string
		index    int
		expectOK bool
	}{
		{"fecha válida con hora", 0, true},
		{"fecha válida simple", 1, true},
		{"formato inválido", 2, false},
		{"vacío", 3, false},
		{"espacios", 4, false},
		{"fuera de rango", 6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDateAt(arr, tt.index)
			if tt.expectOK && got.IsZero() {
				t.Errorf("esperado fecha válida, obtenido zero")
			}
			if !tt.expectOK && !got.IsZero() {
				t.Errorf("esperado zero, obtenido '%v'", got)
			}
		})
	}
}
