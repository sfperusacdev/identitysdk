package agrupador

import "testing"

type ObjA struct {
	id  string
	val int
}

func (o ObjA) ID() string {
	return o.id
}

type ObjB struct {
	id   string
	name string
}

func (o ObjB) ID() string {
	return o.id
}

func TestAgruparAmbosPresentes(t *testing.T) {
	a := []ObjA{
		{id: "1", val: 10},
	}
	b := []ObjB{
		{id: "1", name: "uno"},
	}

	res := Agrupar(a, b)

	if len(res) != 1 {
		t.Fatalf("esperado 1, obtenido %d", len(res))
	}
	if res[0].A == nil || res[0].B == nil {
		t.Fatalf("ambos deben existir")
	}
}

func TestAgruparSoloA(t *testing.T) {
	a := []ObjA{
		{id: "1", val: 10},
	}
	b := []ObjB{}

	res := Agrupar(a, b)

	if len(res) != 1 {
		t.Fatalf("esperado 1, obtenido %d", len(res))
	}
	if res[0].A == nil || res[0].B != nil {
		t.Fatalf("A debe existir y B ser nil")
	}
}

func TestAgruparSoloB(t *testing.T) {
	a := []ObjA{}
	b := []ObjB{
		{id: "1", name: "uno"},
	}

	res := Agrupar(a, b)

	if len(res) != 1 {
		t.Fatalf("esperado 1, obtenido %d", len(res))
	}
	if res[0].A != nil || res[0].B == nil {
		t.Fatalf("A debe ser nil y B existir")
	}
}

func TestAgruparMixto(t *testing.T) {
	a := []ObjA{
		{id: "1", val: 10},
		{id: "2", val: 20},
	}
	b := []ObjB{
		{id: "2", name: "dos"},
		{id: "3", name: "tres"},
	}

	res := Agrupar(a, b)

	if len(res) != 3 {
		t.Fatalf("esperado 3, obtenido %d", len(res))
	}

	m := map[string]Par[ObjA, ObjB]{}
	for _, p := range res {
		if p.A != nil {
			m[p.A.ID()] = p
		} else if p.B != nil {
			m[p.B.ID()] = p
		}
	}

	if m["1"].A == nil || m["1"].B != nil {
		t.Fatalf("id 1 incorrecto")
	}
	if m["2"].A == nil || m["2"].B == nil {
		t.Fatalf("id 2 incorrecto")
	}
	if m["3"].A != nil || m["3"].B == nil {
		t.Fatalf("id 3 incorrecto")
	}
}
