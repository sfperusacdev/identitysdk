package testdb_test

import (
	"context"
	"errors"
	"testing"

	connection "github.com/sfperusacdev/identitysdk/pg-connection"
	"github.com/sfperusacdev/identitysdk/testdb"
	"github.com/stretchr/testify/require"
)

func TestStorageManager_Conn(t *testing.T) {
	ctx := context.Background()
	storage := testdb.NewPostgresStorage(t)

	err := storage.Conn(ctx).Exec(`
		CREATE TABLE conn_test (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		)
	`).Error
	require.NoError(t, err)

	err = storage.Conn(ctx).Exec(
		`INSERT INTO conn_test (name) VALUES (?)`,
		"test",
	).Error
	require.NoError(t, err)

	var count int64
	err = storage.Conn(ctx).Table("conn_test").Count(&count).Error
	require.NoError(t, err)

	require.Equal(t, int64(1), count)
}

func TestStorageManager_WithTx_Commits(t *testing.T) {
	ctx := context.Background()
	storage := testdb.NewPostgresStorage(t)

	err := storage.Conn(ctx).Exec(`
		CREATE TABLE with_tx_commit_test (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		)
	`).Error
	require.NoError(t, err)

	err = storage.WithTx(ctx, func(ctx context.Context) error {
		return storage.Conn(ctx).Exec(
			`INSERT INTO with_tx_commit_test (name) VALUES (?)`,
			"committed",
		).Error
	})
	require.NoError(t, err)

	var count int64
	err = storage.Conn(ctx).Table("with_tx_commit_test").Count(&count).Error
	require.NoError(t, err)

	require.Equal(t, int64(1), count)
}

func TestStorageManager_WithTx_Rollbacks(t *testing.T) {
	ctx := context.Background()
	storage := testdb.NewPostgresStorage(t)

	err := storage.Conn(ctx).Exec(`
		CREATE TABLE with_tx_rollback_test (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		)
	`).Error
	require.NoError(t, err)

	expectedErr := errors.New("force rollback")

	err = storage.WithTx(ctx, func(ctx context.Context) error {
		err := storage.Conn(ctx).Exec(
			`INSERT INTO with_tx_rollback_test (name) VALUES (?)`,
			"rolled-back",
		).Error
		require.NoError(t, err)

		return expectedErr
	})
	require.ErrorIs(t, err, expectedErr)

	var count int64
	err = storage.Conn(ctx).Table("with_tx_rollback_test").Count(&count).Error
	require.NoError(t, err)

	require.Equal(t, int64(0), count)
}

func TestStorageManager_WithTx_NestedTransactionUsesSameTx(t *testing.T) {
	ctx := context.Background()
	storage := testdb.NewPostgresStorage(t)

	err := storage.Conn(ctx).Exec(`
		CREATE TABLE nested_tx_test (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		)
	`).Error
	require.NoError(t, err)

	err = storage.WithTx(ctx, func(ctx context.Context) error {
		err := storage.Conn(ctx).Exec(
			`INSERT INTO nested_tx_test (name) VALUES (?)`,
			"outer",
		).Error
		require.NoError(t, err)

		return storage.WithTx(ctx, func(ctx context.Context) error {
			return storage.Conn(ctx).Exec(
				`INSERT INTO nested_tx_test (name) VALUES (?)`,
				"inner",
			).Error
		})
	})
	require.NoError(t, err)

	var count int64
	err = storage.Conn(ctx).Table("nested_tx_test").Count(&count).Error
	require.NoError(t, err)

	require.Equal(t, int64(2), count)
}

func TestStorageManager_WithTx_NestedTransactionRollsBackAll(t *testing.T) {
	ctx := context.Background()
	storage := testdb.NewPostgresStorage(t)

	err := storage.Conn(ctx).Exec(`
		CREATE TABLE nested_tx_rollback_test (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		)
	`).Error
	require.NoError(t, err)

	expectedErr := errors.New("force nested rollback")

	err = storage.WithTx(ctx, func(ctx context.Context) error {
		err := storage.Conn(ctx).Exec(
			`INSERT INTO nested_tx_rollback_test (name) VALUES (?)`,
			"outer",
		).Error
		require.NoError(t, err)

		err = storage.WithTx(ctx, func(ctx context.Context) error {
			err := storage.Conn(ctx).Exec(
				`INSERT INTO nested_tx_rollback_test (name) VALUES (?)`,
				"inner",
			).Error
			require.NoError(t, err)

			return expectedErr
		})
		require.ErrorIs(t, err, expectedErr)

		return err
	})
	require.ErrorIs(t, err, expectedErr)

	var count int64
	err = storage.Conn(ctx).Table("nested_tx_rollback_test").Count(&count).Error
	require.NoError(t, err)

	require.Equal(t, int64(0), count)
}

func TestStorageManager_WithTx_NilFunc(t *testing.T) {
	ctx := context.Background()
	storage := testdb.NewPostgresStorage(t)

	err := storage.WithTx(ctx, nil)

	require.NoError(t, err)
}

func TestSkipStorage_Conn(t *testing.T) {
	ctx := context.Background()

	storage, err := connection.SkipStorage()
	require.NoError(t, err)

	db := storage.Conn(ctx)

	require.Nil(t, db)
}

func TestSkipStorage_WithTx_ExecutesFunc(t *testing.T) {
	ctx := context.Background()

	storage, err := connection.SkipStorage()
	require.NoError(t, err)

	called := false

	err = storage.WithTx(ctx, func(ctx context.Context) error {
		called = true
		return nil
	})

	require.NoError(t, err)
	require.True(t, called)
}

func TestSkipStorage_WithTx_ReturnsFuncError(t *testing.T) {
	ctx := context.Background()

	storage, err := connection.SkipStorage()
	require.NoError(t, err)

	expectedErr := errors.New("skip storage tx error")

	err = storage.WithTx(ctx, func(ctx context.Context) error {
		return expectedErr
	})

	require.ErrorIs(t, err, expectedErr)
}

func TestSkipStorage_WithTx_NilFunc(t *testing.T) {
	ctx := context.Background()

	storage, err := connection.SkipStorage()
	require.NoError(t, err)

	err = storage.WithTx(ctx, nil)

	require.NoError(t, err)
}
