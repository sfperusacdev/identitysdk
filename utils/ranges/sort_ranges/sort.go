package sortranges

import (
	"slices"

	"github.com/sfperusacdev/identitysdk/utils/ranges"
)

func SortTimeRanges[T ranges.TimeRange](ranges []T) {
	slices.SortFunc(ranges, func(a, b T) int {
		if a.StartTime().Before(b.StartTime()) {
			return -1
		}
		if a.StartTime().After(b.StartTime()) {
			return 1
		}
		// desempate por EndTime
		if a.EndTime().Before(b.EndTime()) {
			return -1
		}
		if a.EndTime().After(b.EndTime()) {
			return 1
		}
		return 0
	})
}
