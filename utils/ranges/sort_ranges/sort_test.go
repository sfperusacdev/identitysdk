package sortranges

import (
	"testing"
	"time"

	"github.com/sfperusacdev/identitysdk/utils/ranges"
)

type testRange struct {
	name  string
	start time.Time
	end   time.Time
}

func (r testRange) StartTime() time.Time { return r.start }
func (r testRange) EndTime() time.Time   { return r.end }

func mustTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestSortTimeRanges_Basic(t *testing.T) {
	ranges := []ranges.TimeRange{
		testRange{"r2", mustTime("2024-01-02T10:00:00Z"), mustTime("2024-01-02T12:00:00Z")},
		testRange{"r1", mustTime("2024-01-01T09:00:00Z"), mustTime("2024-01-01T10:00:00Z")},
		testRange{"r3", mustTime("2024-01-03T08:00:00Z"), mustTime("2024-01-03T09:00:00Z")},
	}

	SortTimeRanges(ranges)

	if ranges[0].(testRange).name != "r1" {
		t.Fatalf("expected r1 first, got %s", ranges[0].(testRange).name)
	}
	if ranges[1].(testRange).name != "r2" {
		t.Fatalf("expected r2 second, got %s", ranges[1].(testRange).name)
	}
	if ranges[2].(testRange).name != "r3" {
		t.Fatalf("expected r3 third, got %s", ranges[2].(testRange).name)
	}
}

func TestSortTimeRanges_TieBreakerEndTime(t *testing.T) {
	ranges := []ranges.TimeRange{
		testRange{"long", mustTime("2024-01-01T09:00:00Z"), mustTime("2024-01-01T12:00:00Z")},
		testRange{"short", mustTime("2024-01-01T09:00:00Z"), mustTime("2024-01-01T10:00:00Z")},
	}

	SortTimeRanges(ranges)

	if ranges[0].(testRange).name != "short" {
		t.Fatalf("expected 'short' first, got %s", ranges[0].(testRange).name)
	}
}

func TestSortTimeRanges_AlreadySorted(t *testing.T) {
	ranges := []ranges.TimeRange{
		testRange{"r1", mustTime("2024-01-01T09:00:00Z"), mustTime("2024-01-01T10:00:00Z")},
		testRange{"r2", mustTime("2024-01-02T09:00:00Z"), mustTime("2024-01-02T10:00:00Z")},
	}

	SortTimeRanges(ranges)

	if ranges[0].(testRange).name != "r1" || ranges[1].(testRange).name != "r2" {
		t.Fatalf("order should remain unchanged")
	}
}

func TestSortTimeRanges_Empty(t *testing.T) {
	var ranges []ranges.TimeRange
	SortTimeRanges(ranges) // no panic esperado
}

func TestSortTimeRanges_Single(t *testing.T) {
	ranges := []ranges.TimeRange{
		testRange{"only", mustTime("2024-01-01T09:00:00Z"), mustTime("2024-01-01T10:00:00Z")},
	}

	SortTimeRanges(ranges)

	if ranges[0].(testRange).name != "only" {
		t.Fatalf("unexpected change")
	}
}
