// Package rangegroups finds groups of time ranges connected by a maximum
// allowed gap.
package rangegroups

import (
	"time"

	"github.com/sfperusacdev/identitysdk/utils/ranges"
)

const maxGroupDuration = 24 * time.Hour

type IdentifiedTimeRange interface {
	ranges.TimeRange
	ID() string
}

func FindGroup[T IdentifiedTimeRange](items []T, id string, maxGap time.Duration) ([]T, bool) {
	startIndex := -1

	for i, item := range items {
		if item.ID() == id {
			startIndex = i
			break
		}
	}

	if startIndex == -1 {
		return nil, false
	}

	left := startIndex
	for left > 0 {
		current := items[left]
		previous := items[left-1]
		gap := current.StartTime().Sub(previous.EndTime())

		if gap > maxGap {
			break
		}

		left--
	}

	right := startIndex
	for right < len(items)-1 {
		current := items[right]
		next := items[right+1]
		gap := next.StartTime().Sub(current.EndTime())

		if gap > maxGap {
			break
		}

		right++
	}

	if items[right].EndTime().Sub(items[left].StartTime()) > maxGroupDuration {
		return nil, false
	}

	return items[left : right+1], true
}
