package searchwindow

import (
	"time"

	"github.com/sfperusacdev/identitysdk/utils/ranges"
)

// SearchContained returns records whose full time range is contained
// inside at least one window.
//
// Borders are inclusive.
// Example:
//
//	window: [10:00 - 12:00]
//	record: [10:00 - 12:00] => included
//	record: [09:30 - 10:30] => excluded
func SearchContained[T ranges.TimeRange, K ranges.TimeWindowsRange](records []T, windows []K) []T {
	return search(records, windows, containsRecord)
}

// SearchOverlapping returns records whose time range overlaps/touches
// at least one window.
//
// Borders are inclusive.
// Example:
//
//	window: [10:00 - 12:00]
//	record: [09:30 - 10:30] => included
//	record: [12:00 - 13:00] => included
//	record: [12:01 - 13:00] => excluded
func SearchOverlapping[T ranges.TimeRange, K ranges.TimeWindowsRange](records []T, windows []K) []T {
	return search(records, windows, overlapsRecord)
}

func search[T ranges.TimeRange, K ranges.TimeWindowsRange](
	records []T,
	windows []K,
	matches func(recordStart, recordEnd, windowStart, windowEnd time.Time) bool,
) []T {
	result := make([]T, 0, len(records))

	for _, record := range records {
		recordStart := record.StartTime()
		recordEnd := record.EndTime()

		if !isValidRange(recordStart, recordEnd) {
			continue
		}

		for _, window := range windows {
			windowStart := window.StartWindowsTime()
			windowEnd := window.EndWindowsTime()

			if !isValidRange(windowStart, windowEnd) {
				continue
			}

			if matches(recordStart, recordEnd, windowStart, windowEnd) {
				result = append(result, record)
				break
			}
		}
	}

	return result
}

func isValidRange(start time.Time, end time.Time) bool {
	return !end.Before(start)
}

func containsRecord(
	recordStart time.Time,
	recordEnd time.Time,
	windowStart time.Time,
	windowEnd time.Time,
) bool {
	return !recordStart.Before(windowStart) && !recordEnd.After(windowEnd)
}

func overlapsRecord(
	recordStart time.Time,
	recordEnd time.Time,
	windowStart time.Time,
	windowEnd time.Time,
) bool {
	return !recordEnd.Before(windowStart) && !recordStart.After(windowEnd)
}
