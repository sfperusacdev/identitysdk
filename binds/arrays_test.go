package binds_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk/binds"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/user0608/goones/types"
)

func newContext(body string) echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func mustUUID(s string) uuid.UUID {
	return uuid.MustParse(s)
}

func TestRequestUUIDs(t *testing.T) {
	u1 := "220a4e82-0dc0-46bf-98f9-e0e3c06b2bf7"
	u2 := "ced02bc8-d1c7-4bcc-8eaa-a6ad58df251a"

	tests := []struct {
		name    string
		body    string
		want    types.UUIDArray
		wantErr bool
	}{
		{
			name: "code takes priority over codes",
			body: `{"code":"` + u1 + `","codes":["` + u2 + `"]}`,
			want: types.UUIDArray{mustUUID(u1)},
		},
		{
			name: "codes used when code missing",
			body: `{"codes":["` + u1 + `","` + u2 + `"]}`,
			want: types.UUIDArray{mustUUID(u1), mustUUID(u2)},
		},
		{
			name: "uuid used when others missing",
			body: `{"uuid":"` + u1 + `"}`,
			want: types.UUIDArray{mustUUID(u1)},
		},
		{
			name: "multiple fields uses first match",
			body: `{"codes":["` + u1 + `"],"uuid":"` + u2 + `"}`,
			want: types.UUIDArray{mustUUID(u1)},
		},
		{
			name: "unique values",
			body: `{"codes":["` + u1 + `","` + u1 + `","` + u2 + `"]}`,
			want: types.UUIDArray{mustUUID(u1), mustUUID(u2)},
		},
		{
			name:    "invalid uuid",
			body:    `{"code":"invalid"}`,
			wantErr: true,
		},
		{
			name:    "missing field",
			body:    `{}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newContext(tt.body)

			res, err := binds.RequestUUIDs(c)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}

func TestRequestStrings(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    types.StrArray
		wantErr bool
	}{
		{
			name: "code priority",
			body: `{"code":["a"],"values":["b"]}`,
			want: types.StrArray{"a"},
		},
		{
			name: "values used",
			body: `{"values":["a","b",""]}`,
			want: types.StrArray{"a", "b"},
		},
		{
			name: "multiple fields uses first match",
			body: `{"strings":["x"],"values":["y"]}`,
			want: types.StrArray{"y"},
		},
		{
			name:    "invalid type",
			body:    `{"values":[1,2]}`,
			wantErr: true,
		},
		{
			name:    "missing field",
			body:    `{}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newContext(tt.body)

			res, err := binds.RequestStrings(c)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}
