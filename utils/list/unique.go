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
