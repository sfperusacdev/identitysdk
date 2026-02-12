package usecase

import (
	"context"

	"github.com/sfperusacdev/identitysdk/setup/sqlsyncdata/descriptor"
	"github.com/sfperusacdev/identitysdk/setup/sqlsyncdata/repos"
	"github.com/sfperusacdev/identitysdk/utils/list"
	"github.com/user0608/goones/errs"
)

type TableDescriptors []descriptor.TableDescriptor

type SQLTableUsecase struct {
	descriptors []descriptor.TableDescriptor
	repository  *repos.SQLTableRepository
}

func NewSQLTableUsecase(descriptors TableDescriptors, repository *repos.SQLTableRepository) (*SQLTableUsecase, error) {
	if len(descriptors) > 0 {
		// TODO; validar descriptors
	}
	return &SQLTableUsecase{
		descriptors: descriptors,
		repository:  repository,
	}, nil
}

type TableInfoResponse struct {
	TableName     string `json:"table_name"`
	Script        string `json:"script"`
	StartSync     int64  `json:"start_sync"`
	RetentionDays uint   `json:"retention_days"`
	ReadyOnly     bool   `json:"read_only"`
}

func (s *SQLTableUsecase) getDescriptor(tableName string) (*descriptor.TableDescriptor, error) {
	for _, d := range s.descriptors {
		if d.Table == tableName {
			return &d, nil
		}
	}
	return nil, errs.BadRequestf("la tabla [%s] no existe o no está registrada", tableName)
}

func (s *SQLTableUsecase) GetTablesStatement(ctx context.Context, tables []string) ([]TableInfoResponse, error) {
	tables = list.NonZeroUniques(tables)
	if len(tables) == 0 {
		return []TableInfoResponse{}, nil
	}
	var tablescript = make([]TableInfoResponse, 0, len(tables))
	for _, table := range tables {
		desc, err := s.getDescriptor(table)
		if err != nil {
			return nil, err
		}
		columns, err := s.repository.GetTableColumns(ctx, table)
		if err != nil {
			return nil, err
		}
		var isReadyOnly = desc.IsReadyOnly(columns)
		tablescript = append(tablescript, TableInfoResponse{
			TableName:     table,
			Script:        desc.BuildCreateTableStatement(columns),
			StartSync:     desc.StartSyncAt().UnixMilli(),
			RetentionDays: desc.SinceDays,
			ReadyOnly:     isReadyOnly,
		})
	}
	return tablescript, nil
}
