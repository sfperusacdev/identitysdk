package testdb

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
	connection "github.com/sfperusacdev/identitysdk/pg-connection"
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

	sharedStorage   connection.StorageManager
	sharedContainer *tcpostgres.PostgresContainer
	sharedErr       error
)

var migrationFS fs.FS

func SetMigrationFS(fsys fs.FS) {
	migrationFS = fsys
}

func NewPostgresStorage(t *testing.T, additionalFileSystems ...fs.FS) connection.StorageManager {
	t.Helper()

	postgresOnce.Do(func() {
		sharedStorage, sharedContainer, sharedErr = startPostgres()
	})

	require.NoError(t, sharedErr)
	require.NotNil(t, sharedStorage)

	require.NoError(t, runMigrations(sharedStorage))

	t.Cleanup(func() {
		dropPublicTables(t, sharedStorage)
	})

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

func runMigrations(storage connection.StorageManager) error {
	goose.SetBaseFS(migrationFS)
	var ctx = context.TODO()
	tx := storage.Conn(ctx)
	db, err := tx.DB()
	if err != nil {
		return err
	}
	if migrationFS != nil {
		if err := goose.UpContext(ctx, db, "migrations"); err != nil {
			slog.Error(fmt.Sprintf("Database %s failed", "up"), "error", err)
			return err
		}
	}

	return nil
}

func dropPublicTables(t *testing.T, storage connection.StorageManager) {
	t.Helper()

	err := storage.Conn(context.Background()).Exec(`
		DO $$
		DECLARE
			table_record RECORD;
			schema_record RECORD;
		BEGIN
			FOR table_record IN (
				SELECT tablename
				FROM pg_tables
				WHERE schemaname = 'public'
			) LOOP
				EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(table_record.tablename) || ' CASCADE';
			END LOOP;

			FOR schema_record IN (
				SELECT schema_name
				FROM information_schema.schemata
				WHERE schema_name NOT IN ('public', 'information_schema')
				  AND schema_name NOT LIKE 'pg_%'
			) LOOP
				EXECUTE 'DROP SCHEMA IF EXISTS ' || quote_ident(schema_record.schema_name) || ' CASCADE';
			END LOOP;
		END $$;
	`).Error
	require.NoError(t, err)
}
