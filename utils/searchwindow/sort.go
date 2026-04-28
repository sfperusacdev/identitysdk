package searchwindow

import "slices"

func SortTimeRanges(ranges []TimeRange) {
	slices.SortFunc(ranges, func(a, b TimeRange) int {
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
