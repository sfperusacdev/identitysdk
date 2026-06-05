package usecase_test

import (
	"context"
	"testing"

	"github.com/sfperusacdev/identitysdk/setup/sqlsyncdata/repos"
	"github.com/sfperusacdev/identitysdk/setup/sqlsyncdata/usecase"
	"github.com/sfperusacdev/identitysdk/testdb"
	"github.com/stretchr/testify/require"
)

func TestSQLTableUsecase_SyncTable_WriteOnlyDoesNotReturnServerRows(t *testing.T) {
	ctx := context.Background()
	storage := testdb.NewPostgresStorage(t)

	err := storage.Conn(ctx).Exec(`
		CREATE TABLE write_only_items (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			sync_at BIGINT NOT NULL
		)
	`).Error
	require.NoError(t, err)

	err = storage.Conn(ctx).Exec(
		`INSERT INTO write_only_items (id, name, sync_at) VALUES (?, ?, ?)`,
		"acme.server",
		"server row",
		int64(1),
	).Error
	require.NoError(t, err)

	tableUsecase, err := usecase.NewSQLTableUsecase(
		usecase.TableDescriptors{
			{
				Table:     "write_only_items",
				Columns:   []string{"name"},
				WriteOnly: true,
			},
		},
		repos.NewSQLTableRepository(storage),
	)
	require.NoError(t, err)

	info, err := tableUsecase.GetTablesStatement(ctx, []string{"write_only_items"})
	require.NoError(t, err)
	require.Len(t, info, 1)
	require.True(t, info[0].WriteOnly)
	require.False(t, info[0].ReadyOnly)

	res, err := tableUsecase.SyncTable(ctx, "acme", usecase.TableSyncRequest{
		TableName: "write_only_items",
		SyncAt:    0,
		Payload: []map[string]any{
			{
				"id":   "acme.client",
				"name": "client row",
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"id"}, res.PrimaryKes)
	require.Empty(t, res.Payload)

	var count int64
	err = storage.Conn(ctx).Table("write_only_items").Count(&count).Error
	require.NoError(t, err)
	require.Equal(t, int64(2), count)

	var saved struct {
		Name   string
		SyncAt int64
	}
	err = storage.Conn(ctx).
		Table("write_only_items").
		Select("name", "sync_at").
		Where("id = ?", "acme.client").
		Scan(&saved).Error
	require.NoError(t, err)
	require.Equal(t, "client row", saved.Name)
	require.Positive(t, saved.SyncAt)
}
