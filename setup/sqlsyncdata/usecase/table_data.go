package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

type TableSyncRequest struct {
	TableName string           `json:"table_name"`
	SyncAt    int64            `json:"sync_at"`
	Payload   []map[string]any `json:"payload"`
}

type TableSyncResponse struct {
	PrimaryKes []string         `json:"identifiers"`
	Payload    []map[string]any `json:"payload"`
}

func (s *SQLTableUsecase) composePrimaryKey(tableName string, primaryKeys []string, record map[string]any) string {
	var keyBuilder strings.Builder

	for _, pk := range primaryKeys {
		value, exists := record[pk]
		if !exists {
			slog.Error("failed to compose primary key",
				"table", tableName,
				"primary_keys", strings.Join(primaryKeys, ","),
				"missing_key", pk,
			)
			return ""
		}

		switch v := value.(type) {
		case string:
			keyBuilder.WriteString(v)
		case int:
			keyBuilder.WriteString(strconv.Itoa(v))
		case int64:
			keyBuilder.WriteString(strconv.FormatInt(v, 10))
		case float64:
			keyBuilder.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
		case bool:
			keyBuilder.WriteString(strconv.FormatBool(v))
		default:
			fmt.Fprintf(&keyBuilder, "%v", v)
		}
	}

	return keyBuilder.String()
}

func (s *SQLTableUsecase) SyncTable(ctx context.Context, domain string, req TableSyncRequest) (*TableSyncResponse, error) {
	nowMillis := time.Now().UnixMilli()

	primaryKeys, err := s.repository.GetTablePrimaryKeys(ctx, req.TableName)
	if err != nil {
		return nil, err
	}

	descriptor, err := s.getDescriptor(req.TableName)
	if err != nil {
		return nil, err
	}

	existingRows, err := s.repository.GetTableData(ctx, domain, *descriptor, req.SyncAt)
	if err != nil {
		return nil, err
	}

	incomingKeySet := make(map[string]struct{}, len(req.Payload))
	for i := range req.Payload {
		id := s.composePrimaryKey(req.TableName, primaryKeys, req.Payload[i])
		incomingKeySet[id] = struct{}{}

		req.Payload[i]["sync_at"] = nowMillis

		if err := descriptor.ValidateScope(req.Payload[i], primaryKeys, domain); err != nil {
			return nil, err
		}
	}

	if err := s.repository.InsertData(ctx, req.TableName, req.Payload); err != nil {
		return nil, err
	}

	rowsToReturn := make([]map[string]any, 0, len(existingRows))
	for _, row := range existingRows {
		id := s.composePrimaryKey(req.TableName, primaryKeys, row)
		if _, exists := incomingKeySet[id]; exists {
			continue
		}
		rowsToReturn = append(rowsToReturn, row)
	}

	return &TableSyncResponse{
		PrimaryKes: primaryKeys,
		Payload:    rowsToReturn,
	}, nil
}
