package ranges

import "time"

type TimeRange interface {
	StartTime() time.Time
	EndTime() time.Time
}

type TimeWindowsRange interface {
	StartWindowsTime() time.Time
	EndWindowsTime() time.Time
}
