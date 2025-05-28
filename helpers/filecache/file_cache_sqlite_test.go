package filecache

import (
	"database/sql"
	"testing"
	"time"
)

func createFileStore(t *testing.T) *FileCache {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS file_access (
		filename TEXT PRIMARY KEY,
		last_read TIMESTAMP NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_last_read ON file_access(last_read);
	`
	if _, err := db.Exec(createTableSQL); err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	return &FileCache{db: db}
}

func TestFileAccessDBMethods(t *testing.T) {
	c := createFileStore(t)
	defer c.db.Close()
	// Test upsertFileAccess
	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, f := range files {
		if err := c.upsertFileAccess(f); err != nil {
			t.Fatalf("upsert failed for %s: %v", f, err)
		}
		time.Sleep(10 * time.Millisecond) // ensure different timestamps
	}

	// Check that all files exist
	rows, err := c.db.Query(`SELECT filename FROM file_access`)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	var stored []string
	for rows.Next() {
		var fn string
		if err := rows.Scan(&fn); err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		stored = append(stored, fn)
	}
	if len(stored) != len(files) {
		t.Fatalf("expected %d entries, got %d", len(files), len(stored))
	}

	// Test getOldestAccessedFiles with maxEntries = 2 (should evict oldest 1)
	c.maxEntries = 2
	oldest, err := c.getOldestAccessedFiles()
	if err != nil {
		t.Fatalf("getOldestAccessedFiles failed: %v", err)
	}
	if len(oldest) != 1 {
		t.Fatalf("expected 1 oldest file, got %d", len(oldest))
	}

	// Test deleteFileAccessRecord
	if err := c.deleteFileAccessRecord(oldest[0]); err != nil {
		t.Fatalf("deleteFileAccessRecord failed: %v", err)
	}

	// Confirm deletion
	var count int
	err = c.db.QueryRow(`SELECT COUNT(*) FROM file_access WHERE filename = ?`, oldest[0]).Scan(&count)
	if err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected file_access record deleted")
	}
}

func Test_evict(t *testing.T) {
	c := createFileStore(t)
	defer c.db.Close()
	c.maxEntries = 1

	c.upsertFileAccess("file01")
	time.Sleep(2 * time.Millisecond)
	c.upsertFileAccess("file02")
	time.Sleep(2 * time.Millisecond)
	c.upsertFileAccess("file03")

	c.evict()

	rows, err := c.db.Query(`SELECT filename FROM file_access`)
	if err != nil {
		t.Error(err)
	}
	defer rows.Close()

	var filenames []string
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			t.Error(err)
		}
		filenames = append(filenames, filename)
	}

	if err := rows.Err(); err != nil {
		t.Error(err)
	}
	if len(filenames) != 1 {
		t.Errorf("expected 1 filename, got %d", len(filenames))
	}
	if filenames[0] != "file03" {
		t.Errorf("expected filename 'file03', got '%s'", filenames[0])
	}
}
