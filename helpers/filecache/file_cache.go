package filecache

import (
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	_ "modernc.org/sqlite"
)

var ErrFileNotExists = errors.New("file does not exist in cache")

type FileCache struct {
	m          sync.RWMutex
	baseDir    string
	pathsMap   map[string]struct{}
	maxEntries int
	db         *sql.DB
}

// NewFileCache creates a new file cache in the specified baseDir.
// The minimum allowed eviction interval is 1 hour.
// The minimum allowed maxEntries is 50.
func NewFileCache(baseDir string, maxEntries int, evictInterval time.Duration) (*FileCache, error) {
	dbName := "filecache_metadata.db"
	dbPath := filepath.Join(baseDir, dbName)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS file_access (
		filename TEXT PRIMARY KEY,
		last_read TIMESTAMP NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_last_read ON file_access(last_read);
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	cache := &FileCache{
		baseDir:    baseDir,
		pathsMap:   make(map[string]struct{}),
		maxEntries: max(maxEntries, 50),
		db:         db,
	}

	allPaths, err := GetAllFilePaths(baseDir)
	if err != nil {
		return nil, err
	}
	for _, path := range allPaths {
		filename := filepath.Base(path)
		if filename == dbName {
			continue
		}
		cache.insertFileAccessIfNew(filename)
		cache.pathsMap[path] = struct{}{}
	}
	go cache.startEvictionLoop(max(evictInterval, 1*time.Hour))
	return cache, nil
}

func (c *FileCache) startEvictionLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			c.evict()
		case <-signalCh:
			return
		}
	}
}

func (c *FileCache) insertFileAccessIfNew(filename string) error {
	now := time.Now()
	_, err := c.db.Exec(`
		INSERT INTO file_access (filename, last_read) VALUES (?, ?)
		ON CONFLICT(filename) DO NOTHING
	`, filename, now)
	return err
}

// upsertFileAccess inserts or updates the last_read timestamp
// for the given filename in the file_access table.
func (c *FileCache) upsertFileAccess(filename string) error {
	now := time.Now()
	_, err := c.db.Exec(`
		INSERT INTO file_access (filename, last_read) VALUES (?, ?)
		ON CONFLICT(filename) DO UPDATE SET last_read=excluded.last_read
	`, filename, now)
	return err
}

// getOldestAccessedFiles returns the list of filenames
// that have the oldest last_read timestamps, excluding
// the newest maxEntries files. It only returns results
// if the total number of entries exceeds maxEntries;
// otherwise, it returns an empty slice.
func (c *FileCache) getOldestAccessedFiles() ([]string, error) {
	var total int
	if err := c.db.QueryRow(`SELECT COUNT(*) FROM file_access`).Scan(&total); err != nil {
		return nil, err
	}
	if total <= c.maxEntries {
		return nil, nil
	}

	rows, err := c.db.Query(`
		SELECT filename FROM file_access
		ORDER BY last_read ASC
		LIMIT ?
	`, total-c.maxEntries)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []string
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		files = append(files, filename)
	}
	return files, nil
}

func (c *FileCache) deleteFileAccessRecord(filename string) error {
	_, err := c.db.Exec(`DELETE FROM file_access WHERE filename = ?`, filename)
	return err
}

func (c *FileCache) exists(filename string) (string, bool) {
	hashed := GetHashedFilePath(filename)
	c.m.RLock()
	defer c.m.RUnlock()
	_, found := c.pathsMap[hashed]
	return filepath.Join(c.baseDir, hashed), found
}

func (c *FileCache) addToCache(filename string) {
	hashed := GetHashedFilePath(filename)
	c.m.Lock()
	defer c.m.Unlock()
	c.pathsMap[hashed] = struct{}{}
}

func (c *FileCache) removeFromCache(filename string) {
	hashed := GetHashedFilePath(filename)
	c.m.Lock()
	defer c.m.Unlock()
	delete(c.pathsMap, hashed)
}

func (c *FileCache) Read(filename string) ([]byte, error) {
	path, ok := c.exists(filename)
	if !ok {
		return nil, ErrFileNotExists
	}
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		c.removeFromCache(filename)
		return nil, ErrFileNotExists
	}
	if err != nil {
		slog.Error("failed to read file", "path", path, "error", err)
		return nil, err
	}
	c.upsertFileAccess(filename)
	return content, nil
}

func (c *FileCache) Write(filename string, fileData []byte) error {
	hashed := GetHashedFilePath(filename)
	fullPath := filepath.Join(c.baseDir, hashed)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Error("failed to create directories", "dir", dir, "error", err)
		return err
	}

	if err := os.WriteFile(fullPath, fileData, 0644); err != nil {
		slog.Error("failed to write file", "path", fullPath, "error", err)
		return err
	}
	c.addToCache(filename)
	c.upsertFileAccess(filename)
	return nil
}

func (c *FileCache) evict() {
	entries, err := c.getOldestAccessedFiles()
	if err != nil {
		slog.Error("failed to get oldest accessed files", "error", err)
		return
	}
	for _, filename := range entries {
		hashed := GetHashedFilePath(filename)
		fullPath := filepath.Join(c.baseDir, hashed)
		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			slog.Error("failed to remove file during eviction", "file", fullPath, "error", err)
			continue
		}
		c.removeFromCache(filename)
		if err := c.deleteFileAccessRecord(filename); err != nil {
			slog.Error("failed to delete file_access record", "filename", filename, "error", err)
		}
	}
}
