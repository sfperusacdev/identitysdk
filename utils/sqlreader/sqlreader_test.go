package sqlreader_test

import (
	"reflect"
	"sort"
	"testing"
	"testing/fstest"

	"github.com/sfperusacdev/identitysdk/utils/sqlreader"
)

func TestLoadSQLFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"migrations/001_init.sql": {
			Data: []byte("CREATE TABLE users(id INT);"),
		},
		"migrations/002_seed.SQL": {
			Data: []byte("INSERT INTO users(id) VALUES (1);"),
		},
		"migrations/readme.txt": {
			Data: []byte("ignore"),
		},
		"migrations/nested/003_more.sql": {
			Data: []byte("ALTER TABLE users ADD COLUMN name TEXT;"),
		},
	}

	got, err := sqlreader.LoadSQLFiles(fsys, "migrations")
	if err != nil {
		t.Fatalf("LoadSQLFiles() error = %v", err)
	}

	sort.Slice(got, func(i, j int) bool {
		return got[i].Path < got[j].Path
	})

	want := []sqlreader.SQLFile{
		{
			Path:    "migrations/001_init.sql",
			Name:    "001_init.sql",
			Content: "CREATE TABLE users(id INT);",
		},
		{
			Path:    "migrations/002_seed.SQL",
			Name:    "002_seed.SQL",
			Content: "INSERT INTO users(id) VALUES (1);",
		},
		{
			Path:    "migrations/nested/003_more.sql",
			Name:    "003_more.sql",
			Content: "ALTER TABLE users ADD COLUMN name TEXT;",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadSQLFiles() = %#v, want %#v", got, want)
	}
}
