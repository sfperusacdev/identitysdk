package staging

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sfperusacdev/identitysdk/configs"
	"go.uber.org/fx"
)

type testLifecycle struct {
	hooks []fx.Hook
}

func (l *testLifecycle) Append(h fx.Hook) {
	l.hooks = append(l.hooks, h)
}

type testConfig struct {
	stagingDir string
}

func (c *testConfig) ListenAddress() string       { return "" }
func (c *testConfig) GRPCAddress() string         { return "" }
func (c *testConfig) Identity() string            { return "" }
func (c *testConfig) IdentityAccessToken() string { return "" }
func (c *testConfig) CacheDir() string            { return "" }
func (c *testConfig) StagingDir() string          { return c.stagingDir }

var _ configs.GeneralServiceConfigProvider = (*testConfig)(nil)

func TestNewStagingFilesAreaLifecycle(t *testing.T) {
	baseDir := t.TempDir()
	lc := &testLifecycle{}
	svc := NewStagingFilesArea(lc, &testConfig{stagingDir: baseDir})
	if svc == nil {
		t.Fatal("expected service")
	}
	if len(lc.hooks) != 1 {
		t.Fatalf("expected 1 lifecycle hook, got %d", len(lc.hooks))
	}
	if err := lc.hooks[0].OnStart(context.Background()); err != nil {
		t.Fatalf("on start failed: %v", err)
	}
	if _, err := os.Stat(baseDir); err != nil {
		t.Fatalf("staging dir was not created: %v", err)
	}
	if err := lc.hooks[0].OnStop(context.Background()); err != nil {
		t.Fatalf("on stop failed: %v", err)
	}
}

func TestWriteReadAndExists(t *testing.T) {
	baseDir := t.TempDir()
	svc := &StagingFilesArea{baseDir: baseDir, cleanupAfter: time.Hour}

	data := []byte("hello world")
	hash, err := svc.Write(data)
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	want := sha1.Sum(data)
	if hash != hex.EncodeToString(want[:]) {
		t.Fatalf("unexpected hash: got %s", hash)
	}

	filePath := filepath.Join(baseDir, hash)
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("stored file missing: %v", err)
	}

	if !svc.Exists(hash) {
		t.Fatal("expected file to exist")
	}

	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(filePath, old, old); err != nil {
		t.Fatalf("failed to set old times: %v", err)
	}
	if !svc.Exists(hash) {
		t.Fatal("expected file to exist after exists check")
	}
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if !info.ModTime().After(old) {
		t.Fatalf("expected mod time to be refreshed, got %v <= %v", info.ModTime(), old)
	}

	if err := os.Chtimes(filePath, old, old); err != nil {
		t.Fatalf("failed to reset old times: %v", err)
	}

	readData, err := svc.Read(hash)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(readData) != string(data) {
		t.Fatalf("unexpected read content: got %q", readData)
	}
	info, err = os.Stat(filePath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if !info.ModTime().After(old) {
		t.Fatalf("expected mod time to be refreshed after read, got %v <= %v", info.ModTime(), old)
	}
}

func TestReadAndExistsRejectInvalidHash(t *testing.T) {
	baseDir := t.TempDir()
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "outside")
	if err := os.WriteFile(outsideFile, []byte("outside"), 0644); err != nil {
		t.Fatalf("failed to write outside file: %v", err)
	}

	svc := &StagingFilesArea{baseDir: baseDir, cleanupAfter: time.Hour}
	invalidHashes := []string{
		"",
		"short",
		"../" + filepath.Base(outsideFile),
		filepath.Join("..", filepath.Base(outsideDir), filepath.Base(outsideFile)),
		"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
		"00112233445566778899aabbccddeeff0011223/",
	}

	for _, hash := range invalidHashes {
		t.Run(hash, func(t *testing.T) {
			if svc.Exists(hash) {
				t.Fatal("expected invalid hash to not exist")
			}
			if _, err := svc.Read(hash); !errors.Is(err, ErrInvalidHash) {
				t.Fatalf("expected ErrInvalidHash, got %v", err)
			}
		})
	}
}

func TestCleanupInterval(t *testing.T) {
	tests := []struct {
		name         string
		cleanupAfter time.Duration
		want         time.Duration
	}{
		{"small duration", time.Minute, time.Minute},
		{"half threshold", 30 * time.Minute, 15 * time.Minute},
		{"max interval", 24 * time.Hour, time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanupInterval(tt.cleanupAfter); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCleanRemovesOldFiles(t *testing.T) {
	baseDir := t.TempDir()
	svc := &StagingFilesArea{baseDir: baseDir, cleanupAfter: time.Hour}

	oldFile := filepath.Join(baseDir, "old")
	newFile := filepath.Join(baseDir, "new")
	if err := os.WriteFile(oldFile, []byte("old"), 0644); err != nil {
		t.Fatalf("failed to write old file: %v", err)
	}
	if err := os.WriteFile(newFile, []byte("new"), 0644); err != nil {
		t.Fatalf("failed to write new file: %v", err)
	}

	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set old file time: %v", err)
	}

	if err := svc.clean(); err != nil {
		t.Fatalf("clean failed: %v", err)
	}

	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Fatalf("expected old file to be removed, got err=%v", err)
	}
	if _, err := os.Stat(newFile); err != nil {
		t.Fatalf("expected new file to remain: %v", err)
	}
}
