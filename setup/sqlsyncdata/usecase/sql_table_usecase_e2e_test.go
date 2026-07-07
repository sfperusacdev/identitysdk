package usecase_test

import (
	"context"
	"fmt"
	"testing"

	connection "github.com/sfperusacdev/identitysdk/pg-connection"
	"github.com/sfperusacdev/identitysdk/setup/sqlsyncdata/repos"
	"github.com/sfperusacdev/identitysdk/setup/sqlsyncdata/usecase"
	"github.com/sfperusacdev/identitysdk/testdb"
	"github.com/stretchr/testify/require"
)

func TestSQLTableUsecase_SyncTable_ReturnsIncrementalRowsAndPersistsPayload(t *testing.T) {
	ctx := context.Background()
	storage := testdb.NewPostgresStorage(t)
	createSyncItemsTable(t, ctx, storage, "sync_items")
	insertSyncItem(t, ctx, storage, "sync_items", "acme.old", "old row", 10)
	insertSyncItem(t, ctx, storage, "sync_items", "acme.server", "server row", 200)
	insertSyncItem(t, ctx, storage, "sync_items", "other.server", "other domain", 200)

	tableUsecase := newTableUsecase(t, storage, usecase.TableDescriptors{
		{Table: "sync_items", Columns: []string{"name"}},
	})

	res, err := tableUsecase.SyncTable(ctx, "acme", usecase.TableSyncRequest{
		TableName: "sync_items",
		SyncAt:    100,
		Payload: []map[string]any{
			{"id": "acme.client", "name": "client row"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"id"}, res.PrimaryKes)
	require.Len(t, res.Payload, 1)
	require.Equal(t, "acme.server", res.Payload[0]["id"])
	require.Equal(t, "server row", res.Payload[0]["name"])

	var saved syncItemRow
	err = storage.Conn(ctx).
		Table("sync_items").
		Select("id", "name", "sync_at").
		Where("id = ?", "acme.client").
		Scan(&saved).Error
	require.NoError(t, err)
	require.Equal(t, "acme.client", saved.ID)
	require.Equal(t, "client row", saved.Name)
	require.Positive(t, saved.SyncAt)
}

func TestSQLTableUsecase_SyncTable_DoesNotReturnIncomingRows(t *testing.T) {
	ctx := context.Background()
	storage := testdb.NewPostgresStorage(t)
	createSyncItemsTable(t, ctx, storage, "sync_dedup_items")
	insertSyncItem(t, ctx, storage, "sync_dedup_items", "acme.item", "server version", 200)

	tableUsecase := newTableUsecase(t, storage, usecase.TableDescriptors{
		{Table: "sync_dedup_items", Columns: []string{"name"}},
	})

	res, err := tableUsecase.SyncTable(ctx, "acme", usecase.TableSyncRequest{
		TableName: "sync_dedup_items",
		SyncAt:    100,
		Payload: []map[string]any{
			{"id": "acme.item", "name": "client version"},
		},
	})
	require.NoError(t, err)
	require.Empty(t, res.Payload)

	var saved syncItemRow
	err = storage.Conn(ctx).
		Table("sync_dedup_items").
		Select("id", "name", "sync_at").
		Where("id = ?", "acme.item").
		Scan(&saved).Error
	require.NoError(t, err)
	require.Equal(t, "client version", saved.Name)
	require.Greater(t, saved.SyncAt, int64(200))
}

func TestSQLTableUsecase_SyncTable_WriteOnlyDoesNotReturnServerRows(t *testing.T) {
	ctx := context.Background()
	storage := testdb.NewPostgresStorage(t)
	createSyncItemsTable(t, ctx, storage, "write_only_items")
	insertSyncItem(t, ctx, storage, "write_only_items", "acme.server", "server row", 1)

	tableUsecase := newTableUsecase(t, storage, usecase.TableDescriptors{
		{Table: "write_only_items", Columns: []string{"name"}, WriteOnly: true},
	})

	info, err := tableUsecase.GetTablesStatement(ctx, []string{"write_only_items"})
	require.NoError(t, err)
	require.Len(t, info, 1)
	require.True(t, info[0].WriteOnly)
	require.False(t, info[0].ReadyOnly)

	res, err := tableUsecase.SyncTable(ctx, "acme", usecase.TableSyncRequest{
		TableName: "write_only_items",
		SyncAt:    0,
		Payload: []map[string]any{
			{"id": "acme.client", "name": "client row"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"id"}, res.PrimaryKes)
	require.Empty(t, res.Payload)

	var count int64
	err = storage.Conn(ctx).Table("write_only_items").Count(&count).Error
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}

func TestSQLTableUsecase_SyncTable_ReadOnlyBlocksPayload(t *testing.T) {
	ctx := context.Background()
	storage := testdb.NewPostgresStorage(t)
	createSyncItemsTable(t, ctx, storage, "read_only_items")

	tableUsecase := newTableUsecase(t, storage, usecase.TableDescriptors{
		{Table: "read_only_items", Columns: []string{"name"}, ReadOnly: true},
	})

	_, err := tableUsecase.SyncTable(ctx, "acme", usecase.TableSyncRequest{
		TableName: "read_only_items",
		SyncAt:    0,
		Payload: []map[string]any{
			{"id": "acme.client", "name": "client row"},
		},
	})
	require.Error(t, err)

	var count int64
	err = storage.Conn(ctx).Table("read_only_items").Count(&count).Error
	require.NoError(t, err)
	require.Equal(t, int64(0), count)
}

func TestSQLTableUsecase_SyncTable_ReadOnlyAllowsPull(t *testing.T) {
	ctx := context.Background()
	storage := testdb.NewPostgresStorage(t)
	createSyncItemsTable(t, ctx, storage, "read_only_pull_items")
	insertSyncItem(t, ctx, storage, "read_only_pull_items", "acme.server", "server row", 200)

	tableUsecase := newTableUsecase(t, storage, usecase.TableDescriptors{
		{Table: "read_only_pull_items", Columns: []string{"name"}, ReadOnly: true},
	})

	res, err := tableUsecase.SyncTable(ctx, "acme", usecase.TableSyncRequest{
		TableName: "read_only_pull_items",
		SyncAt:    100,
	})
	require.NoError(t, err)
	require.Len(t, res.Payload, 1)
	require.Equal(t, "acme.server", res.Payload[0]["id"])
}

func TestSQLTableUsecase_SyncTable_ReadOnlyUsesConfiguredPrimaryKeys(t *testing.T) {
	ctx := context.Background()
	storage := testdb.NewPostgresStorage(t)
	createReadOnlyItemsTable(t, ctx, storage, "readonly_configured_pk_items")
	insertSyncItem(t, ctx, storage, "readonly_configured_pk_items", "acme.server", "server row", 200)

	tableUsecase := newTableUsecase(t, storage, usecase.TableDescriptors{
		{Table: "readonly_configured_pk_items", Columns: []string{"name"}, ReadOnly: true, PrimaryKeys: []string{"id"}},
	})

	res, err := tableUsecase.SyncTable(ctx, "acme", usecase.TableSyncRequest{
		TableName: "readonly_configured_pk_items",
		SyncAt:    100,
	})
	require.NoError(t, err)
	require.Equal(t, []string{"id"}, res.PrimaryKes)
	require.Len(t, res.Payload, 1)
	require.Equal(t, "acme.server", res.Payload[0]["id"])
}

func TestSQLTableUsecase_SyncTable_RejectsConfiguredPrimaryKeysOnWritableTable(t *testing.T) {
	ctx := context.Background()
	storage := testdb.NewPostgresStorage(t)
	createSyncItemsTable(t, ctx, storage, "writable_configured_pk_items")

	tableUsecase := newTableUsecase(t, storage, usecase.TableDescriptors{
		{Table: "writable_configured_pk_items", Columns: []string{"name"}, PrimaryKeys: []string{"id"}},
	})

	_, err := tableUsecase.SyncTable(ctx, "acme", usecase.TableSyncRequest{
		TableName: "writable_configured_pk_items",
		SyncAt:    100,
	})
	require.Error(t, err)
}

func TestSQLTableUsecase_SyncTable_RejectsPayloadOutsideDomain(t *testing.T) {
	ctx := context.Background()
	storage := testdb.NewPostgresStorage(t)
	createSyncItemsTable(t, ctx, storage, "scope_items")

	tableUsecase := newTableUsecase(t, storage, usecase.TableDescriptors{
		{Table: "scope_items", Columns: []string{"name"}},
	})

	_, err := tableUsecase.SyncTable(ctx, "acme", usecase.TableSyncRequest{
		TableName: "scope_items",
		SyncAt:    0,
		Payload: []map[string]any{
			{"id": "other.client", "name": "client row"},
		},
	})
	require.Error(t, err)

	var count int64
	err = storage.Conn(ctx).Table("scope_items").Count(&count).Error
	require.NoError(t, err)
	require.Equal(t, int64(0), count)
}

type syncItemRow struct {
	ID     string `gorm:"column:id"`
	Name   string `gorm:"column:name"`
	SyncAt int64  `gorm:"column:sync_at"`
}

func newTableUsecase(
	t *testing.T,
	storage connection.StorageManager,
	descriptors usecase.TableDescriptors,
) *usecase.SQLTableUsecase {
	t.Helper()

	tableUsecase, err := usecase.NewSQLTableUsecase(
		descriptors,
		repos.NewSQLTableRepository(storage),
	)
	require.NoError(t, err)
	return tableUsecase
}

func createSyncItemsTable(t *testing.T, ctx context.Context, storage connection.StorageManager, table string) {
	t.Helper()

	err := storage.Conn(ctx).Exec(fmt.Sprintf(`
		CREATE TABLE %s (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			sync_at BIGINT NOT NULL
		)
	`, table)).Error
	require.NoError(t, err)
}

func createReadOnlyItemsTable(t *testing.T, ctx context.Context, storage connection.StorageManager, table string) {
	t.Helper()

	err := storage.Conn(ctx).Exec(fmt.Sprintf(`
		CREATE TABLE %s (
			id TEXT NOT NULL,
			name TEXT NOT NULL,
			sync_at BIGINT NOT NULL
		)
	`, table)).Error
	require.NoError(t, err)
}

func insertSyncItem(
	t *testing.T,
	ctx context.Context,
	storage connection.StorageManager,
	table string,
	id string,
	name string,
	syncAt int64,
) {
	t.Helper()

	err := storage.Conn(ctx).Exec(
		fmt.Sprintf(`INSERT INTO %s (id, name, sync_at) VALUES (?, ?, ?)`, table),
		id,
		name,
		syncAt,
	).Error
	require.NoError(t, err)
}
