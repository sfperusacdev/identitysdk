package staging

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sfperusacdev/identitysdk/configs"
	"go.uber.org/fx"
)

type StagingFilesArea struct {
	baseDir      string
	cleanupAfter time.Duration
}

const defaultCleanupAfter = 24 * time.Hour

var ErrInvalidHash = errors.New("invalid hash")

func NewStagingFilesArea(lc fx.Lifecycle, config configs.GeneralServiceConfigProvider) *StagingFilesArea {

	staging := &StagingFilesArea{
		baseDir:      config.StagingDir(),
		cleanupAfter: defaultCleanupAfter,
	}
	var (
		mu     sync.Mutex
		cancel context.CancelFunc
	)

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			if err := os.MkdirAll(staging.baseDir, 0755); err != nil {
				return err
			}
			if err := staging.clean(); err != nil {
				return err
			}

			ctx, c := context.WithCancel(context.Background())
			mu.Lock()
			cancel = c
			mu.Unlock()

			go staging.runCleaner(ctx)
			return nil
		},
		OnStop: func(_ context.Context) error {
			mu.Lock()
			c := cancel
			cancel = nil
			mu.Unlock()
			if c != nil {
				c()
			}
			return nil
		},
	})

	return staging
}

func (s *StagingFilesArea) Write(data []byte) (string, error) {
	if err := os.MkdirAll(s.baseDir, 0755); err != nil {
		slog.Error("failed to create staging directory", "dir", s.baseDir, "error", err)
		return "", err
	}

	hash := hashBytes(data)
	finalPath := s.filePath(hash)
	tempFile, err := os.CreateTemp(s.baseDir, hash+".tmp-*")
	if err != nil {
		slog.Error("failed to create temp file", "dir", s.baseDir, "error", err)
		return "", err
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = os.Remove(tempPath)
	}()

	if _, err := tempFile.Write(data); err != nil {
		slog.Error("failed to write temp file", "path", tempPath, "error", err)
		_ = tempFile.Close()
		return "", err
	}
	if err := tempFile.Sync(); err != nil {
		slog.Error("failed to sync temp file", "path", tempPath, "error", err)
		_ = tempFile.Close()
		return "", err
	}
	if err := tempFile.Close(); err != nil {
		slog.Error("failed to close temp file", "path", tempPath, "error", err)
		return "", err
	}
	if err := os.Rename(tempPath, finalPath); err != nil {
		slog.Error("failed to rename temp file", "from", tempPath, "to", finalPath, "error", err)
		return "", err
	}
	_ = os.Chtimes(finalPath, time.Now(), time.Now())
	return hash, nil
}

func (s *StagingFilesArea) Exists(hash string) bool {
	if !isValidHash(hash) {
		return false
	}

	filePath := s.filePath(hash)
	if _, err := os.Stat(filePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			slog.Error("failed to stat staging file", "path", filePath, "error", err)
		}
		return false
	}
	if err := os.Chtimes(filePath, time.Now(), time.Now()); err != nil {
		slog.Error("failed to refresh staging file time", "path", filePath, "error", err)
	}
	return true
}

func (s *StagingFilesArea) Read(hash string) ([]byte, error) {
	if !isValidHash(hash) {
		return nil, ErrInvalidHash
	}

	filePath := s.filePath(hash)
	data, err := os.ReadFile(filePath)
	if err != nil {
		slog.Error("failed to read staging file", "path", filePath, "error", err)
		return nil, err
	}
	if err := os.Chtimes(filePath, time.Now(), time.Now()); err != nil {
		slog.Error("failed to refresh staging file time", "path", filePath, "error", err)
	}
	return data, nil
}

func (s *StagingFilesArea) clean() error {
	cutoff := time.Now().Add(-s.cleanupAfter)
	return filepath.WalkDir(s.baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Error("failed to walk staging directory", "path", path, "error", err)
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			slog.Error("failed to read staging file info", "path", path, "error", err)
			return err
		}
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				slog.Error("failed to remove stale staging file", "path", path, "error", err)
				return err
			}
		}
		return nil
	})
}

func (s *StagingFilesArea) runCleaner(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval(s.cleanupAfter))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = s.clean()
		case <-ctx.Done():
			return
		}
	}
}

func (s *StagingFilesArea) filePath(hash string) string {
	return filepath.Join(s.baseDir, hash)
}

func cleanupInterval(cleanupAfter time.Duration) time.Duration {
	const maxInterval = time.Hour
	if cleanupAfter <= 2*time.Minute {
		return cleanupAfter
	}
	interval := cleanupAfter / 2
	if interval > maxInterval {
		return maxInterval
	}
	return interval
}

func isValidHash(hash string) bool {
	if len(hash) != sha1.Size*2 {
		return false
	}
	for _, r := range hash {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}

func hashBytes(data []byte) string {
	h := sha1.Sum(data)
	return hex.EncodeToString(h[:])
}
