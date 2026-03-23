package formfile_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sfperusacdev/identitysdk/binds/formfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMultipartContext(t *testing.T, fieldName string, content []byte) echo.Context {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if fieldName != "" {
		part, err := writer.CreateFormFile(fieldName, "test.txt")
		require.NoError(t, err)
		_, err = part.Write(content)
		require.NoError(t, err)
	}

	writer.Close()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()

	return e.NewContext(req, rec)
}

func TestReadOptionalFormFile(t *testing.T) {
	t.Run("file present", func(t *testing.T) {
		c := newMultipartContext(t, "file", []byte("hello"))

		data, err := formfile.ReadOptionalFormFile(c, "file")

		require.NoError(t, err)
		assert.Equal(t, []byte("hello"), data)
	})

	t.Run("file missing", func(t *testing.T) {
		c := newMultipartContext(t, "", nil)

		data, err := formfile.ReadOptionalFormFile(c, "file")

		require.NoError(t, err)
		assert.Nil(t, data)
	})
}

func TestReadRequiredFormFile(t *testing.T) {
	t.Run("file present", func(t *testing.T) {
		c := newMultipartContext(t, "file", []byte("hello"))

		data, err := formfile.ReadRequiredFormFile(c, "file")

		require.NoError(t, err)
		assert.Equal(t, []byte("hello"), data)
	})

	t.Run("file missing", func(t *testing.T) {
		c := newMultipartContext(t, "", nil)

		data, err := formfile.ReadRequiredFormFile(c, "file")

		assert.Error(t, err)
		assert.Nil(t, data)
	})
}
