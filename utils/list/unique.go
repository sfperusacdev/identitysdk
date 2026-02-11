package list

func Unique[T comparable](list []T) []T {
	seen := make(map[T]struct{}, len(list))
	out := make([]T, 0, len(list))

	for _, v := range list {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}

	return out
}

// NonzeroUniques returns only the nonzero unique values from a slice.
func NonZeroUniques[T comparable](list []T) []T {
	result := make([]T, 0, len(list))
	existMap := make(map[T]struct{}, len(list))

	var zeroVal T

	for _, val := range list {
		if val == zeroVal {
			continue
		}
		if _, ok := existMap[val]; ok {
			continue
		}
		existMap[val] = struct{}{}
		result = append(result, val)
	}

	return result
}
