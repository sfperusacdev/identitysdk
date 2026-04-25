package workallocation

type TimeRangeOverlap[T TimeRange] struct {
	Range       T
	Overlapping bool
}

func MarkOverlappingRanges[T TimeRange](ranges []T) []TimeRangeOverlap[T] {
	result := make([]TimeRangeOverlap[T], len(ranges))

	for i, r := range ranges {
		result[i] = TimeRangeOverlap[T]{
			Range:       r,
			Overlapping: false,
		}
	}

	for i := range ranges {
		for j := i + 1; j < len(ranges); j++ {
			if ranges[i].StartTime().Before(ranges[j].EndTime()) &&
				ranges[j].StartTime().Before(ranges[i].EndTime()) {
				result[i].Overlapping = true
				result[j].Overlapping = true
			}
		}
	}

	return result
}
