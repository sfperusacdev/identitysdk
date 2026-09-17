package list

import (
	"reflect"
	"testing"
)

func TestResolveByValue_SearchableFields(t *testing.T) {
	type item struct {
		Codigo      string
		Code        string
		ID          string
		UUID        string
		Descripcion string
		Description string
		Name        string
		Nombre      string
	}

	tests := []struct {
		name  string
		item  item
		query string
	}{
		{name: "codigo", item: item{Codigo: "codigo-1"}, query: "codigo-1"},
		{name: "code", item: item{Code: "code-1"}, query: "code-1"},
		{name: "id", item: item{ID: "id-1"}, query: "id-1"},
		{name: "uuid", item: item{UUID: "uuid-1"}, query: "uuid-1"},
		{name: "descripcion", item: item{Descripcion: "descripcion-1"}, query: "descripcion-1"},
		{name: "description", item: item{Description: "description-1"}, query: "description-1"},
		{name: "name", item: item{Name: "name-1"}, query: "name-1"},
		{name: "nombre", item: item{Nombre: "nombre-1"}, query: "nombre-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := ResolveByValue([]item{tt.item}, tt.query)
			if resolved == nil || !reflect.DeepEqual(*resolved, tt.item) {
				t.Fatalf("got %v, want %v", resolved, tt.item)
			}
		})
	}
}

func TestResolveByValue_FieldPriority(t *testing.T) {
	type item struct {
		Codigo      string
		Description string
		Name        string
	}

	items := []item{
		{Description: "same", Name: "same"},
		{Codigo: "same"},
	}
	resolved := ResolveByValue(items, "same")
	if resolved == nil || resolved.Codigo != "same" {
		t.Fatalf("got %v, want item matched by codigo", resolved)
	}

	items = []item{{Name: "same"}, {Description: "same"}}
	resolved = ResolveByValue(items, "same")
	if resolved == nil || resolved.Description != "same" {
		t.Fatalf("got %v, want item matched by description", resolved)
	}
}

func TestResolveByValue_JSONTags(t *testing.T) {
	type item struct {
		Reference string `json:"codigo"`
		Label     string `json:"nombre,omitempty"`
	}

	items := []item{{Reference: "C-1"}, {Label: "Etiqueta"}}
	if resolved := ResolveByValue(items, "c-1"); resolved == nil || resolved.Reference != "C-1" {
		t.Fatalf("got %v, want item with codigo C-1", resolved)
	}
	if resolved := ResolveByValue(items, "etiqueta"); resolved == nil || resolved.Label != "Etiqueta" {
		t.Fatalf("got %v, want item with nombre Etiqueta", resolved)
	}
}

func TestResolveByValue_Normalization(t *testing.T) {
	type item struct {
		Name string
	}

	tests := []struct {
		name  string
		value string
		query string
	}{
		{name: "uppercase", value: "Código", query: "CODIGO"},
		{name: "accented vowels", value: "áéíóú", query: "aeiou"},
		{name: "diaeresis", value: "pingüino", query: "PINGUINO"},
		{name: "enye", value: "niño", query: "NINO"},
		{name: "surrounding spaces", value: "  valor  ", query: " valor "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := ResolveByValue([]item{{Name: tt.value}}, tt.query)
			if resolved == nil || resolved.Name != tt.value {
				t.Fatalf("got %v, want %q", resolved, tt.value)
			}
		})
	}
}

func TestResolveByValue_PrimitivesAndNumbers(t *testing.T) {
	type namedString string
	type namedInt int

	if resolved := ResolveByValue([]string{"uno", "dos"}, " DOS "); resolved == nil || *resolved != "dos" {
		t.Fatalf("got %v, want dos", resolved)
	}
	if resolved := ResolveByValue([]int{10, 20, 30}, "20"); resolved == nil || *resolved != 20 {
		t.Fatalf("got %v, want 20", resolved)
	}
	if resolved := ResolveByValue([]int64{100, 200}, "200"); resolved == nil || *resolved != 200 {
		t.Fatalf("got %v, want 200", resolved)
	}
	if resolved := ResolveByValue([]float64{1.5, 2.5}, "2.5"); resolved == nil || *resolved != 2.5 {
		t.Fatalf("got %v, want 2.5", resolved)
	}
	if resolved := ResolveByValue([]namedString{"alpha"}, "ALPHA"); resolved == nil || *resolved != "alpha" {
		t.Fatalf("got %v, want alpha", resolved)
	}
	if resolved := ResolveByValue([]namedInt{7}, "7"); resolved == nil || *resolved != 7 {
		t.Fatalf("got %v, want 7", resolved)
	}
}

func TestResolveByValue_NumericFields(t *testing.T) {
	type item struct {
		ID          int
		Description float64
	}

	items := []item{{ID: 12, Description: 1.5}, {ID: 123, Description: 2.5}}
	if resolved := ResolveByValue(items, "12"); resolved == nil || resolved.ID != 12 {
		t.Fatalf("got %v, want item with id 12", resolved)
	}
	if resolved := ResolveByValue(items, "2.5"); resolved == nil || resolved.ID != 123 {
		t.Fatalf("got %v, want item with description 2.5", resolved)
	}
	if resolved := ResolveByValue(items, "1"); resolved != nil {
		t.Fatalf("got %v, want nil for partial numeric match", resolved)
	}
}

func TestResolveByValue_PointersAndNilFields(t *testing.T) {
	type item struct {
		ID   *int
		Name *string
	}

	id := 42
	name := "Nombre"
	items := []*item{nil, {}, {ID: &id, Name: &name}}

	if resolved := ResolveByValue(items, "42"); resolved == nil || *(*resolved).ID != 42 {
		t.Fatalf("got %v, want item with id 42", resolved)
	}
	if resolved := ResolveByValue(items, "nombre"); resolved == nil || *(*resolved).Name != "Nombre" {
		t.Fatalf("got %v, want item with name Nombre", resolved)
	}
}

func TestResolveByValue_ExactMatchAndEmptyInput(t *testing.T) {
	items := []string{"1", "12", "123"}
	if resolved := ResolveByValue(items, "1"); resolved == nil || *resolved != "1" {
		t.Fatalf("got %v, want exact value 1", resolved)
	}
	if resolved := ResolveByValue(items, "2"); resolved != nil {
		t.Fatalf("got %v, want nil for non-exact match", resolved)
	}
	if resolved := ResolveByValue([]string{}, "1"); resolved != nil {
		t.Fatalf("got %v, want nil for empty list", resolved)
	}
	if resolved := ResolveByValue(items, " "); resolved != nil {
		t.Fatalf("got %v, want nil for empty search", resolved)
	}
}
