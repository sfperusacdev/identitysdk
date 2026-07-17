// Package schedulegroups estimates representative schedules by grouping time
// ranges whose start and end times are close to each other.
package schedulegroups

import (
	"slices"
	"time"

	"github.com/sfperusacdev/identitysdk/utils/ranges"
	"github.com/user0608/goones/types"
)

const minutesPerDay = 24 * 60

type ScheduleGroup[T ranges.TimeRange] struct {
	Start  types.JustTime
	End    types.JustTime
	Ranges []T
}

type normalizedRange[T ranges.TimeRange] struct {
	start    int
	end      int
	original T
}

type group[T ranges.TimeRange] struct {
	start int
	end   int
	items []normalizedRange[T]
}

func Estimate[T ranges.TimeRange](items []T, margin time.Duration) []ScheduleGroup[T] {
	if len(items) == 0 {
		return nil
	}

	if margin < 0 {
		margin = 0
	}

	marginMinutes := int(margin / time.Minute)
	normalized := normalizeRanges(items)
	slices.SortFunc(normalized, func(a, b normalizedRange[T]) int {
		if a.start < b.start {
			return -1
		}
		if a.start > b.start {
			return 1
		}
		if a.end < b.end {
			return -1
		}
		if a.end > b.end {
			return 1
		}
		return 0
	})

	groups := make([]group[T], 0)
	for _, item := range normalized {
		bestIndex := -1
		bestDistance := 0

		for i, group := range groups {
			startDiff := abs(item.start - group.start)
			endDiff := abs(item.end - group.end)

			if startDiff > marginMinutes || endDiff > marginMinutes {
				continue
			}

			distance := startDiff + endDiff
			if bestIndex == -1 || distance < bestDistance {
				bestIndex = i
				bestDistance = distance
			}
		}

		if bestIndex == -1 {
			groups = append(groups, group[T]{
				start: item.start,
				end:   item.end,
				items: []normalizedRange[T]{item},
			})
			continue
		}

		groups[bestIndex].items = append(groups[bestIndex].items, item)
		recalculateRepresentative(&groups[bestIndex])
	}

	return toScheduleGroups(groups)
}

func normalizeRanges[T ranges.TimeRange](items []T) []normalizedRange[T] {
	result := make([]normalizedRange[T], 0, len(items))

	for _, item := range items {
		start := minutesSinceMidnight(item.StartTime())
		end := minutesSinceMidnight(item.EndTime())
		dayDiff := dateOnly(item.EndTime()).Sub(dateOnly(item.StartTime())) / (24 * time.Hour)

		result = append(result, normalizedRange[T]{
			start:    start,
			end:      end + int(dayDiff)*minutesPerDay,
			original: item,
		})
	}

	return result
}

func recalculateRepresentative[T ranges.TimeRange](group *group[T]) {
	starts := make([]int, 0, len(group.items))
	ends := make([]int, 0, len(group.items))

	for _, item := range group.items {
		starts = append(starts, item.start)
		ends = append(ends, item.end)
	}

	group.start = median(starts)
	group.end = median(ends)
}

func toScheduleGroups[T ranges.TimeRange](groups []group[T]) []ScheduleGroup[T] {
	result := make([]ScheduleGroup[T], 0, len(groups))

	for _, group := range groups {
		scheduleGroup := ScheduleGroup[T]{
			Start:  justTimeFromMinutes(group.start),
			End:    justTimeFromMinutes(group.end),
			Ranges: make([]T, 0, len(group.items)),
		}

		for _, item := range group.items {
			scheduleGroup.Ranges = append(scheduleGroup.Ranges, item.original)
		}

		result = append(result, scheduleGroup)
	}

	return result
}

func median(values []int) int {
	slices.Sort(values)
	middle := len(values) / 2

	if len(values)%2 == 1 {
		return values[middle]
	}

	return (values[middle-1] + values[middle]) / 2
}

func minutesSinceMidnight(t time.Time) int {
	return t.Hour()*60 + t.Minute()
}

func justTimeFromMinutes(minutes int) types.JustTime {
	minutes = ((minutes % minutesPerDay) + minutesPerDay) % minutesPerDay
	return types.JustTime(time.Duration(minutes) * time.Minute)
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func abs(value int) int {
	if value < 0 {
		return -value
	}

	return value
}
