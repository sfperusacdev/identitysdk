package searchwindow

import (
	"reflect"
	"testing"
	"time"
)

type testRecord struct {
	ID    string
	Start time.Time
	End   time.Time
}

func (r testRecord) StartTime() time.Time {
	return r.Start
}

func (r testRecord) EndTime() time.Time {
	return r.End
}

type testWindow struct {
	Start time.Time
	End   time.Time
}

func (w testWindow) StartWindowsTime() time.Time {
	return w.Start
}

func (w testWindow) EndWindowsTime() time.Time {
	return w.End
}

func TestSearchContainedWithDates(t *testing.T) {
	windows := []testWindow{
		{
			Start: datetime(2026, time.April, 29, 10, 0),
			End:   datetime(2026, time.April, 29, 12, 0),
		},
	}

	records := []testRecord{
		{
			ID:    "exact",
			Start: datetime(2026, time.April, 29, 10, 0),
			End:   datetime(2026, time.April, 29, 12, 0),
		},
		{
			ID:    "inside",
			Start: datetime(2026, time.April, 29, 10, 30),
			End:   datetime(2026, time.April, 29, 11, 30),
		},
		{
			ID:    "starts-before",
			Start: datetime(2026, time.April, 29, 9, 30),
			End:   datetime(2026, time.April, 29, 11, 0),
		},
		{
			ID:    "ends-after",
			Start: datetime(2026, time.April, 29, 11, 0),
			End:   datetime(2026, time.April, 29, 12, 30),
		},
		{
			ID:    "before",
			Start: datetime(2026, time.April, 29, 8, 0),
			End:   datetime(2026, time.April, 29, 9, 0),
		},
		{
			ID:    "after",
			Start: datetime(2026, time.April, 29, 13, 0),
			End:   datetime(2026, time.April, 29, 14, 0),
		},
	}

	got := recordIDs(SearchContained(records, windows))
	want := []string{
		"exact",
		"inside",
	}

	assertEqualIDs(t, want, got)
}

func TestSearchOverlappingWithDates(t *testing.T) {
	windows := []testWindow{
		{
			Start: datetime(2026, time.April, 29, 10, 0),
			End:   datetime(2026, time.April, 29, 12, 0),
		},
	}

	records := []testRecord{
		{
			ID:    "exact",
			Start: datetime(2026, time.April, 29, 10, 0),
			End:   datetime(2026, time.April, 29, 12, 0),
		},
		{
			ID:    "inside",
			Start: datetime(2026, time.April, 29, 10, 30),
			End:   datetime(2026, time.April, 29, 11, 30),
		},
		{
			ID:    "starts-before-but-overlaps",
			Start: datetime(2026, time.April, 29, 9, 30),
			End:   datetime(2026, time.April, 29, 10, 30),
		},
		{
			ID:    "ends-after-but-overlaps",
			Start: datetime(2026, time.April, 29, 11, 30),
			End:   datetime(2026, time.April, 29, 12, 30),
		},
		{
			ID:    "touches-start",
			Start: datetime(2026, time.April, 29, 9, 0),
			End:   datetime(2026, time.April, 29, 10, 0),
		},
		{
			ID:    "touches-end",
			Start: datetime(2026, time.April, 29, 12, 0),
			End:   datetime(2026, time.April, 29, 13, 0),
		},
		{
			ID:    "before",
			Start: datetime(2026, time.April, 29, 8, 0),
			End:   datetime(2026, time.April, 29, 9, 59),
		},
		{
			ID:    "after",
			Start: datetime(2026, time.April, 29, 12, 1),
			End:   datetime(2026, time.April, 29, 13, 0),
		},
	}

	got := recordIDs(SearchOverlapping(records, windows))
	want := []string{
		"exact",
		"inside",
		"starts-before-but-overlaps",
		"ends-after-but-overlaps",
		"touches-start",
		"touches-end",
	}

	assertEqualIDs(t, want, got)
}

func TestSearchContainedWithDatesAcrossDifferentDays(t *testing.T) {
	windows := []testWindow{
		{
			Start: datetime(2026, time.April, 29, 22, 0),
			End:   datetime(2026, time.April, 30, 2, 0),
		},
	}

	records := []testRecord{
		{
			ID:    "inside-next-day",
			Start: datetime(2026, time.April, 29, 23, 0),
			End:   datetime(2026, time.April, 30, 1, 0),
		},
		{
			ID:    "exact-next-day",
			Start: datetime(2026, time.April, 29, 22, 0),
			End:   datetime(2026, time.April, 30, 2, 0),
		},
		{
			ID:    "starts-before",
			Start: datetime(2026, time.April, 29, 21, 59),
			End:   datetime(2026, time.April, 30, 1, 0),
		},
		{
			ID:    "ends-after",
			Start: datetime(2026, time.April, 29, 23, 0),
			End:   datetime(2026, time.April, 30, 2, 1),
		},
	}

	got := recordIDs(SearchContained(records, windows))
	want := []string{
		"inside-next-day",
		"exact-next-day",
	}

	assertEqualIDs(t, want, got)
}

func TestSearchOverlappingWithDatesAcrossDifferentDays(t *testing.T) {
	windows := []testWindow{
		{
			Start: datetime(2026, time.April, 29, 22, 0),
			End:   datetime(2026, time.April, 30, 2, 0),
		},
	}

	records := []testRecord{
		{
			ID:    "inside-next-day",
			Start: datetime(2026, time.April, 29, 23, 0),
			End:   datetime(2026, time.April, 30, 1, 0),
		},
		{
			ID:    "starts-before-but-overlaps",
			Start: datetime(2026, time.April, 29, 21, 0),
			End:   datetime(2026, time.April, 29, 22, 30),
		},
		{
			ID:    "ends-after-but-overlaps",
			Start: datetime(2026, time.April, 30, 1, 30),
			End:   datetime(2026, time.April, 30, 3, 0),
		},
		{
			ID:    "before",
			Start: datetime(2026, time.April, 29, 20, 0),
			End:   datetime(2026, time.April, 29, 21, 59),
		},
		{
			ID:    "after",
			Start: datetime(2026, time.April, 30, 2, 1),
			End:   datetime(2026, time.April, 30, 3, 0),
		},
	}

	got := recordIDs(SearchOverlapping(records, windows))
	want := []string{
		"inside-next-day",
		"starts-before-but-overlaps",
		"ends-after-but-overlaps",
	}

	assertEqualIDs(t, want, got)
}

func TestSearchContainedWithHoursOnly(t *testing.T) {
	windows := []testWindow{
		{
			Start: hour(10, 0),
			End:   hour(12, 0),
		},
	}

	records := []testRecord{
		{
			ID:    "exact-hour-range",
			Start: hour(10, 0),
			End:   hour(12, 0),
		},
		{
			ID:    "inside-hour-range",
			Start: hour(10, 30),
			End:   hour(11, 30),
		},
		{
			ID:    "starts-before-hour-window",
			Start: hour(9, 30),
			End:   hour(11, 0),
		},
		{
			ID:    "ends-after-hour-window",
			Start: hour(11, 0),
			End:   hour(12, 30),
		},
		{
			ID:    "before-hour-window",
			Start: hour(8, 0),
			End:   hour(9, 0),
		},
		{
			ID:    "after-hour-window",
			Start: hour(13, 0),
			End:   hour(14, 0),
		},
	}

	got := recordIDs(SearchContained(records, windows))
	want := []string{
		"exact-hour-range",
		"inside-hour-range",
	}

	assertEqualIDs(t, want, got)
}

func TestSearchOverlappingWithHoursOnly(t *testing.T) {
	windows := []testWindow{
		{
			Start: hour(10, 0),
			End:   hour(12, 0),
		},
	}

	records := []testRecord{
		{
			ID:    "exact-hour-range",
			Start: hour(10, 0),
			End:   hour(12, 0),
		},
		{
			ID:    "inside-hour-range",
			Start: hour(10, 30),
			End:   hour(11, 30),
		},
		{
			ID:    "starts-before-but-overlaps-hour-window",
			Start: hour(9, 30),
			End:   hour(10, 30),
		},
		{
			ID:    "ends-after-but-overlaps-hour-window",
			Start: hour(11, 30),
			End:   hour(12, 30),
		},
		{
			ID:    "touches-hour-window-start",
			Start: hour(9, 0),
			End:   hour(10, 0),
		},
		{
			ID:    "touches-hour-window-end",
			Start: hour(12, 0),
			End:   hour(13, 0),
		},
		{
			ID:    "before-hour-window",
			Start: hour(8, 0),
			End:   hour(9, 59),
		},
		{
			ID:    "after-hour-window",
			Start: hour(12, 1),
			End:   hour(13, 0),
		},
	}

	got := recordIDs(SearchOverlapping(records, windows))
	want := []string{
		"exact-hour-range",
		"inside-hour-range",
		"starts-before-but-overlaps-hour-window",
		"ends-after-but-overlaps-hour-window",
		"touches-hour-window-start",
		"touches-hour-window-end",
	}

	assertEqualIDs(t, want, got)
}

func TestSearchDoesNotDuplicateRecordsWhenRecordMatchesMultipleWindows(t *testing.T) {
	windows := []testWindow{
		{
			Start: datetime(2026, time.April, 29, 10, 0),
			End:   datetime(2026, time.April, 29, 12, 0),
		},
		{
			Start: datetime(2026, time.April, 29, 11, 0),
			End:   datetime(2026, time.April, 29, 13, 0),
		},
	}

	records := []testRecord{
		{
			ID:    "matches-two-windows",
			Start: datetime(2026, time.April, 29, 11, 30),
			End:   datetime(2026, time.April, 29, 11, 45),
		},
	}

	got := recordIDs(SearchOverlapping(records, windows))
	want := []string{
		"matches-two-windows",
	}

	assertEqualIDs(t, want, got)
}

func TestSearchIgnoresInvalidRecords(t *testing.T) {
	windows := []testWindow{
		{
			Start: datetime(2026, time.April, 29, 10, 0),
			End:   datetime(2026, time.April, 29, 12, 0),
		},
	}

	records := []testRecord{
		{
			ID:    "invalid-record",
			Start: datetime(2026, time.April, 29, 12, 0),
			End:   datetime(2026, time.April, 29, 10, 0),
		},
		{
			ID:    "valid-record",
			Start: datetime(2026, time.April, 29, 10, 30),
			End:   datetime(2026, time.April, 29, 11, 30),
		},
	}

	got := recordIDs(SearchContained(records, windows))
	want := []string{
		"valid-record",
	}

	assertEqualIDs(t, want, got)
}

func TestSearchIgnoresInvalidWindows(t *testing.T) {
	windows := []testWindow{
		{
			Start: datetime(2026, time.April, 29, 12, 0),
			End:   datetime(2026, time.April, 29, 10, 0),
		},
	}

	records := []testRecord{
		{
			ID:    "valid-record",
			Start: datetime(2026, time.April, 29, 10, 30),
			End:   datetime(2026, time.April, 29, 11, 30),
		},
	}

	got := recordIDs(SearchContained(records, windows))
	want := []string{}

	assertEqualIDs(t, want, got)
}

func datetime(year int, month time.Month, day int, hour int, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}

func hour(hour int, minute int) time.Time {
	return time.Date(0, time.January, 1, hour, minute, 0, 0, time.UTC)
}

func recordIDs(records []testRecord) []string {
	ids := make([]string, 0, len(records))

	for _, record := range records {
		ids = append(ids, record.ID)
	}

	return ids
}

func assertEqualIDs(t *testing.T, want []string, got []string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
