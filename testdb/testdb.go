package testdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
	connection "github.com/sfperusacdev/identitysdk/pg-connection"
	"github.com/sfperusacdev/identitysdk/utils/sql/sqlreader"
	"github.com/sfperusacdev/identitysdk/utils/sql/sqlviews"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	postgresImage    = "postgres:18-alpine"
	testDatabaseName = "testdb"
	testUsername     = "testuser"
	testPassword     = "testpass"
	testLogLevel     = "error"
)

var (
	postgresOnce sync.Once
	testdbMu     sync.Mutex

	sharedStorage   connection.StorageManager
	sharedContainer *tcpostgres.PostgresContainer
	sharedErr       error
)

func NewPostgresStorage(t *testing.T, migrationFS fs.FS) connection.StorageManager {
	t.Helper()
	testdbMu.Lock()

	postgresOnce.Do(func() {
		sharedStorage, sharedContainer, sharedErr = startPostgres()
	})

	t.Cleanup(func() {
		if sharedStorage != nil {
			if err := resetPublicSchema(sharedStorage); err != nil {
				t.Error(err)
			}
		}
		testdbMu.Unlock()
	})

	require.NoError(t, sharedErr)
	require.NotNil(t, sharedStorage)

	require.NoError(t, runMigrations(sharedStorage, migrationFS))

	return sharedStorage
}

func TerminatePostgres(ctx context.Context) error {
	if sharedContainer == nil {
		return nil
	}

	return testcontainers.TerminateContainer(sharedContainer)
}

func startPostgres() (connection.StorageManager, *tcpostgres.PostgresContainer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	container, err := tcpostgres.Run(
		ctx,
		postgresImage,
		tcpostgres.WithDatabase(testDatabaseName),
		tcpostgres.WithUsername(testUsername),
		tcpostgres.WithPassword(testPassword),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, nil, err
	}

	storage, err := newStorageFromContainer(ctx, container)
	if err != nil {
		_ = testcontainers.TerminateContainer(container)
		return nil, nil, err
	}

	return storage, container, nil
}

func newStorageFromContainer(ctx context.Context, container *tcpostgres.PostgresContainer) (connection.StorageManager, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(mappedPort.Port())
	if err != nil {
		return nil, err
	}

	return connection.NewConnection(connection.DBConfigParams{
		DBHost:     host,
		DBPort:     fmt.Sprint(port),
		DBUsername: testUsername,
		DBName:     testDatabaseName,
		DBPassword: testPassword,
		DBLogLevel: testLogLevel,
	})
}

func runMigrations(storage connection.StorageManager, migrationFS fs.FS) error {
	goose.SetBaseFS(migrationFS)
	var ctx = context.TODO()
	tx := storage.Conn(ctx)
	db, err := tx.DB()
	if err != nil {
		return err
	}
	if migrationFS != nil {
		viewFiles, err := loadViewFiles(migrationFS)
		if err != nil {
			return err
		}
		if err := dropViews(db, viewFiles); err != nil {
			return err
		}
		if err := goose.UpContext(ctx, db, "migrations"); err != nil {
			slog.Error(fmt.Sprintf("Database %s failed", "up"), "error", err)
			return err
		}
		if err := recoverViews(db, viewFiles); err != nil {
			return err
		}
	}

	return nil
}

type dbViewFile struct {
	name  string
	sql   string
	views []string
}

func loadViewFiles(fsys fs.FS) ([]dbViewFile, error) {
	files, err := sqlreader.LoadSQLFiles(fsys, "migrations/_views")
	if errors.Is(err, fs.ErrNotExist) {
		return []dbViewFile{}, nil
	}
	if err != nil {
		return nil, err
	}

	result := make([]dbViewFile, 0, len(files))
	for _, file := range files {
		if file.Content == "" {
			continue
		}
		result = append(result, dbViewFile{
			name:  file.Name,
			sql:   file.Content,
			views: sqlviews.FindViewNames(file.Content),
		})
	}
	return result, nil
}

func dropViews(db *sql.DB, files []dbViewFile) error {
	for _, file := range files {
		for _, view := range file.views {
			if _, err := db.Exec("DROP VIEW IF EXISTS " + view); err != nil {
				return err
			}
		}
	}
	return nil
}

func recoverViews(db *sql.DB, files []dbViewFile) error {
	for _, file := range files {
		if _, err := db.Exec(file.sql); err != nil {
			return fmt.Errorf("failed to execute view file %s: %w", file.name, err)
		}
	}
	return nil
}

func resetPublicSchema(storage connection.StorageManager) error {
	return storage.Conn(context.Background()).Exec(`
		DROP SCHEMA IF EXISTS public CASCADE;
		CREATE SCHEMA public;
		GRANT ALL ON SCHEMA public TO public;
	`).Error
}
