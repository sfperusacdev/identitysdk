package binds_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/binds"
	"github.com/sfperusacdev/identitysdk/entities"
	"github.com/stretchr/testify/require"
)

func ctxWithTZ(req *http.Request, loc *time.Location) *http.Request {
	ctx := identitysdk.BuildContext(
		context.Background(),
		"",
		&entities.JwtData{Jwt: entities.Jwt{Zona: loc.String()}},
	)

	return req.WithContext(ctx)
}

func newEchoContext(target string, loc *time.Location) echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, target, nil)

	if loc != nil {
		req = ctxWithTZ(req, loc)
	}

	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func expectedDayRangeUTC(loc *time.Location, y int, m time.Month, d int) (time.Time, time.Time) {
	startLocal := time.Date(y, m, d, 0, 0, 0, 0, loc)
	endLocal := time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), loc)

	return startLocal.UTC(), endLocal.UTC()
}

func TestQueryDateRangeUTC_MultipleTimezones(t *testing.T) {
	zones := []string{
		"America/Lima",
		"UTC",
		"America/New_York",
		"Europe/Madrid",
		"Asia/Tokyo",
	}

	for _, zone := range zones {
		t.Run(zone, func(t *testing.T) {
			loc, err := time.LoadLocation(zone)
			require.NoError(t, err)

			c := newEchoContext("/?desde=2026-04-01&hasta=2026-04-02", loc)

			from, to, err := binds.QueryDateRangeUTC(c)

			require.NoError(t, err)

			expectedFrom, _ := expectedDayRangeUTC(loc, 2026, time.April, 1)
			_, expectedTo := expectedDayRangeUTC(loc, 2026, time.April, 2)

			require.Equal(t, expectedFrom, from)
			require.Equal(t, expectedTo, to)
		})
	}
}

func TestQuerySingleDateRangeUTC_MultipleTimezones(t *testing.T) {
	zones := []string{
		"America/Lima",
		"UTC",
		"America/New_York",
		"Europe/Madrid",
		"Asia/Tokyo",
	}

	for _, zone := range zones {
		t.Run(zone, func(t *testing.T) {
			loc, err := time.LoadLocation(zone)
			require.NoError(t, err)

			c := newEchoContext("/?fecha=2026-04-27", loc)

			start, end, err := binds.QuerySingleDateRangeUTC(c)

			require.NoError(t, err)

			expectedStart, expectedEnd := expectedDayRangeUTC(loc, 2026, time.April, 27)

			require.Equal(t, expectedStart, start)
			require.Equal(t, expectedEnd, end)
		})
	}
}

func TestQueryDateRangeUTC_FromAliases(t *testing.T) {
	loc := time.UTC

	tests := []struct {
		name  string
		param string
	}{
		{name: "desde", param: "desde"},
		{name: "inicio", param: "inicio"},
		{name: "from", param: "from"},
		{name: "start", param: "start"},
		{name: "left", param: "left"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newEchoContext("/?"+tt.param+"=2026-04-01&hasta=2026-04-02", loc)

			from, to, err := binds.QueryDateRangeUTC(c)

			require.NoError(t, err)

			expectedFrom, _ := expectedDayRangeUTC(loc, 2026, time.April, 1)
			_, expectedTo := expectedDayRangeUTC(loc, 2026, time.April, 2)

			require.Equal(t, expectedFrom, from)
			require.Equal(t, expectedTo, to)
		})
	}
}

func TestQueryDateRangeUTC_ToAliases(t *testing.T) {
	loc := time.UTC

	tests := []struct {
		name  string
		param string
	}{
		{name: "hasta", param: "hasta"},
		{name: "fin", param: "fin"},
		{name: "to", param: "to"},
		{name: "end", param: "end"},
		{name: "right", param: "right"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newEchoContext("/?desde=2026-04-01&"+tt.param+"=2026-04-02", loc)

			from, to, err := binds.QueryDateRangeUTC(c)

			require.NoError(t, err)

			expectedFrom, _ := expectedDayRangeUTC(loc, 2026, time.April, 1)
			_, expectedTo := expectedDayRangeUTC(loc, 2026, time.April, 2)

			require.Equal(t, expectedFrom, from)
			require.Equal(t, expectedTo, to)
		})
	}
}

func TestQuerySingleDateRangeUTC_Aliases(t *testing.T) {
	loc := time.UTC

	tests := []struct {
		name  string
		param string
	}{
		{name: "fecha", param: "fecha"},
		{name: "date", param: "date"},
		{name: "dia", param: "dia"},
		{name: "day", param: "day"},
		{name: "on", param: "on"},
		{name: "at", param: "at"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newEchoContext("/?"+tt.param+"=2026-04-27", loc)

			start, end, err := binds.QuerySingleDateRangeUTC(c)

			require.NoError(t, err)

			expectedStart, expectedEnd := expectedDayRangeUTC(loc, 2026, time.April, 27)

			require.Equal(t, expectedStart, start)
			require.Equal(t, expectedEnd, end)
		})
	}
}

func TestQueryDateRangeUTC_ParamPriority(t *testing.T) {
	loc := time.UTC

	t.Run("from priority uses desde before from", func(t *testing.T) {
		c := newEchoContext("/?desde=2026-04-01&from=2026-05-01&hasta=2026-04-02", loc)

		from, _, err := binds.QueryDateRangeUTC(c)

		require.NoError(t, err)

		expectedFrom, _ := expectedDayRangeUTC(loc, 2026, time.April, 1)
		require.Equal(t, expectedFrom, from)
	})

	t.Run("to priority uses hasta before to", func(t *testing.T) {
		c := newEchoContext("/?desde=2026-04-01&hasta=2026-04-02&to=2026-05-02", loc)

		_, to, err := binds.QueryDateRangeUTC(c)

		require.NoError(t, err)

		_, expectedTo := expectedDayRangeUTC(loc, 2026, time.April, 2)
		require.Equal(t, expectedTo, to)
	})
}

func TestQuerySingleDateRangeUTC_ParamPriority(t *testing.T) {
	loc := time.UTC

	c := newEchoContext("/?fecha=2026-04-27&date=2026-05-27", loc)

	start, end, err := binds.QuerySingleDateRangeUTC(c)

	require.NoError(t, err)

	expectedStart, expectedEnd := expectedDayRangeUTC(loc, 2026, time.April, 27)

	require.Equal(t, expectedStart, start)
	require.Equal(t, expectedEnd, end)
}

func TestQueryDateRangeUTC_Errors(t *testing.T) {
	loc := time.UTC

	tests := []struct {
		name   string
		target string
	}{
		{
			name:   "invalid from date",
			target: "/?desde=bad&hasta=2026-04-02",
		},
		{
			name:   "invalid to date",
			target: "/?desde=2026-04-01&hasta=bad",
		},
		{
			name:   "missing from date",
			target: "/?hasta=2026-04-02",
		},
		{
			name:   "missing to date",
			target: "/?desde=2026-04-01",
		},
		{
			name:   "missing both dates",
			target: "/",
		},
		{
			name:   "empty from date",
			target: "/?desde=&hasta=2026-04-02",
		},
		{
			name:   "empty to date",
			target: "/?desde=2026-04-01&hasta=",
		},
		{
			name:   "datetime is not date only",
			target: "/?desde=2026-04-01T10:00:00Z&hasta=2026-04-02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newEchoContext(tt.target, loc)

			from, to, err := binds.QueryDateRangeUTC(c)

			require.Error(t, err)
			require.True(t, from.IsZero())
			require.True(t, to.IsZero())
		})
	}
}

func TestQuerySingleDateRangeUTC_Errors(t *testing.T) {
	loc := time.UTC

	tests := []struct {
		name   string
		target string
	}{
		{
			name:   "invalid date",
			target: "/?fecha=bad",
		},
		{
			name:   "missing date",
			target: "/",
		},
		{
			name:   "empty date",
			target: "/?fecha=",
		},
		{
			name:   "datetime is not date only",
			target: "/?fecha=2026-04-27T10:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newEchoContext(tt.target, loc)

			start, end, err := binds.QuerySingleDateRangeUTC(c)

			require.Error(t, err)
			require.True(t, start.IsZero())
			require.True(t, end.IsZero())
		})
	}
}

func TestQueryDateRangeUTC_MissingTimezone(t *testing.T) {
	c := newEchoContext("/?desde=2026-04-01&hasta=2026-04-02", nil)

	from, to, err := binds.QueryDateRangeUTC(c)

	require.Error(t, err)
	require.True(t, from.IsZero())
	require.True(t, to.IsZero())
}

func TestQuerySingleDateRangeUTC_MissingTimezone(t *testing.T) {
	c := newEchoContext("/?fecha=2026-04-27", nil)

	start, end, err := binds.QuerySingleDateRangeUTC(c)

	require.Error(t, err)
	require.True(t, start.IsZero())
	require.True(t, end.IsZero())
}
