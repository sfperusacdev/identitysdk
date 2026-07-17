package schedulegroups

import (
	"reflect"
	"testing"
	"time"

	"github.com/user0608/goones/types"
)

type testRange struct {
	id    string
	start time.Time
	end   time.Time
}

func (r testRange) StartTime() time.Time { return r.start }
func (r testRange) EndTime() time.Time   { return r.end }

func datetime(day int, hour int, minute int) time.Time {
	return time.Date(2026, 7, day, hour, minute, 0, 0, time.UTC)
}

func jt(hour int, minute int) types.JustTime {
	return types.JustTime(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)
}

func TestEstimateWithExample(t *testing.T) {
	items := []testRange{
		{id: "a", start: datetime(10, 7, 55), end: datetime(10, 16, 55)},
		{id: "b", start: datetime(11, 8, 5), end: datetime(11, 17, 5)},
		{id: "c", start: datetime(12, 8, 15), end: datetime(12, 17, 0)},
		{id: "d", start: datetime(13, 9, 50), end: datetime(13, 16, 55)},
		{id: "e", start: datetime(14, 10, 5), end: datetime(14, 17, 5)},
		{id: "f", start: datetime(15, 10, 10), end: datetime(15, 17, 10)},
		{id: "g", start: datetime(16, 13, 15), end: datetime(16, 19, 30)},
	}

	groups := Estimate(items, 30*time.Minute)

	assertGroupCount(t, groups, 3)
	assertTotalRanges(t, groups, len(items))
	assertGroup(t, groups[0], jt(8, 5), jt(17, 0), []string{"a", "b", "c"})
	assertGroup(t, groups[1], jt(10, 5), jt(17, 5), []string{"d", "e", "f"})
	assertGroup(t, groups[2], jt(13, 15), jt(19, 30), []string{"g"})
}

func TestEstimateValidatesStartAndEndMarginsSeparately(t *testing.T) {
	items := []testRange{
		{id: "a", start: datetime(10, 8, 0), end: datetime(10, 17, 0)},
		{id: "b", start: datetime(11, 8, 10), end: datetime(11, 18, 0)},
	}

	groups := Estimate(items, 30*time.Minute)

	assertGroupCount(t, groups, 2)
	assertGroup(t, groups[0], jt(8, 0), jt(17, 0), []string{"a"})
	assertGroup(t, groups[1], jt(8, 10), jt(18, 0), []string{"b"})
}

func TestEstimateChoosesClosestMatchingGroup(t *testing.T) {
	items := []testRange{
		{id: "a", start: datetime(10, 8, 0), end: datetime(10, 17, 0)},
		{id: "b", start: datetime(11, 8, 50), end: datetime(11, 18, 30)},
		{id: "c", start: datetime(12, 9, 0), end: datetime(12, 17, 55)},
	}

	groups := Estimate(items, time.Hour)

	assertGroupCount(t, groups, 2)
	assertGroup(t, groups[0], jt(8, 0), jt(17, 0), []string{"a"})
	assertGroup(t, groups[1], jt(8, 55), jt(18, 12), []string{"b", "c"})
}

func TestEstimateNightShift(t *testing.T) {
	items := []testRange{
		{id: "a", start: datetime(10, 22, 0), end: datetime(11, 6, 0)},
		{id: "b", start: datetime(12, 22, 10), end: datetime(13, 6, 5)},
	}

	groups := Estimate(items, 15*time.Minute)

	assertGroupCount(t, groups, 1)
	assertGroup(t, groups[0], jt(22, 5), jt(6, 2), []string{"a", "b"})
}

func TestEstimateNegativeMarginUsesZero(t *testing.T) {
	items := []testRange{
		{id: "a", start: datetime(10, 8, 0), end: datetime(10, 17, 0)},
		{id: "b", start: datetime(11, 8, 0), end: datetime(11, 17, 0)},
		{id: "c", start: datetime(12, 8, 1), end: datetime(12, 17, 0)},
	}

	groups := Estimate(items, -time.Minute)

	assertGroupCount(t, groups, 2)
	assertGroup(t, groups[0], jt(8, 0), jt(17, 0), []string{"a", "b"})
	assertGroup(t, groups[1], jt(8, 1), jt(17, 0), []string{"c"})
}

func TestEstimateEmpty(t *testing.T) {
	if groups := Estimate([]testRange{}, 30*time.Minute); groups != nil {
		t.Fatalf("expected nil, got %v", groups)
	}
}

func assertGroupCount(t *testing.T, groups []ScheduleGroup[testRange], want int) {
	t.Helper()

	if len(groups) != want {
		t.Fatalf("expected %d groups, got %d", want, len(groups))
	}
}

func assertTotalRanges(t *testing.T, groups []ScheduleGroup[testRange], want int) {
	t.Helper()

	total := 0
	for _, group := range groups {
		total += len(group.Ranges)
	}

	if total != want {
		t.Fatalf("expected %d assigned ranges, got %d", want, total)
	}
}

func assertGroup(
	t *testing.T,
	group ScheduleGroup[testRange],
	wantStart types.JustTime,
	wantEnd types.JustTime,
	wantIDs []string,
) {
	t.Helper()

	if group.Start != wantStart {
		t.Fatalf("expected start %s, got %s", wantStart, group.Start)
	}

	if group.End != wantEnd {
		t.Fatalf("expected end %s, got %s", wantEnd, group.End)
	}

	gotIDs := rangeIDs(group.Ranges)
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("expected range IDs %v, got %v", wantIDs, gotIDs)
	}
}

func rangeIDs(items []testRange) []string {
	ids := make([]string, 0, len(items))

	for _, item := range items {
		ids = append(ids, item.id)
	}

	return ids
}
