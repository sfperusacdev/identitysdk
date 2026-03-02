package agrupador

type Identificable interface {
	ID() string
}

type Par[A any, B any] struct {
	A *A
	B *B
}

func Agrupar[A Identificable, B Identificable](listaA []A, listaB []B) []Par[A, B] {
	mapaA := make(map[string]A)
	for _, a := range listaA {
		mapaA[a.ID()] = a
	}

	mapaB := make(map[string]B)
	for _, b := range listaB {
		mapaB[b.ID()] = b
	}

	ids := make(map[string]struct{})
	for id := range mapaA {
		ids[id] = struct{}{}
	}
	for id := range mapaB {
		ids[id] = struct{}{}
	}

	resultado := make([]Par[A, B], 0, len(ids))
	for id := range ids {
		var pa *A
		var pb *B

		if a, ok := mapaA[id]; ok {
			aa := a
			pa = &aa
		}
		if b, ok := mapaB[id]; ok {
			bb := b
			pb = &bb
		}

		resultado = append(resultado, Par[A, B]{A: pa, B: pb})
	}

	return resultado
}
