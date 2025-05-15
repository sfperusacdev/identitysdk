package propsprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/sfperusacdev/identitysdk"

	"github.com/sfperusacdev/identitysdk/helpers/properties"
	"github.com/sfperusacdev/identitysdk/helpers/properties/models"
	connection "github.com/sfperusacdev/identitysdk/pg-connection"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
	"github.com/user0608/goones/errs"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type PropItem struct {
	Key   string `gorm:"primaryKey;size:255" json:"key"`
	Value string `gorm:"type:text;not null" json:"value"`
}

func (*PropItem) TableName() string { return "_system_properties" }

type SystemPropsPgProvider struct {
	ready   sync.Map
	manager connection.StorageManager
	entries []models.DetailedSystemProperty
}

var _ properties.SystemPropsProvider = (*SystemPropsPgProvider)(nil)
var _ properties.SystemPropertiesMutator = (*SystemPropsPgProvider)(nil)

func NewSystemPropsPgProvider(
	manager connection.StorageManager,
	entries []models.DetailedSystemProperty,
) *SystemPropsPgProvider {
	return &SystemPropsPgProvider{manager: manager, entries: entries}
}

func (r *SystemPropsPgProvider) ensureTable(ctx context.Context, empresa string) error {
	if _, exists := r.ready.Load(empresa); exists {
		return nil
	}
	const script = `
	CREATE TABLE IF NOT EXISTS _system_properties (
		key VARCHAR(255) PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		grupo TEXT,
		description TEXT,
		value TEXT NOT NULL,
		priority SMALLINT DEFAULT 0 CHECK (priority >= 0),
		data_type TEXT NOT NULL CHECK (data_type IN ('string', 'number', 'boolean', 'array', 'object'))
	)`
	var tx = r.manager.Conn(ctx)
	rs := tx.Session(&gorm.Session{Logger: logger.Discard}).Exec(script)
	if rs.Error != nil {
		return errs.Pgf(rs.Error)
	}
	var records = make([]map[string]any, 0, len(r.entries))
	for i, entry := range r.entries {
		records = append(
			records,
			map[string]any{
				"key":         identitysdk.Empresa(ctx, entry.ID),
				"title":       entry.Title,
				"grupo":       entry.Group,
				"description": entry.Description,
				"value":       entry.Value,
				"priority":    i + 1,
				"data_type":   entry.Type,
			},
		)
	}
	if len(records) > 0 {
		rs = tx.Session(&gorm.Session{Logger: logger.Discard}).
			Clauses(
				clause.OnConflict{
					Columns:   []clause.Column{{Name: "key"}},
					DoUpdates: clause.AssignmentColumns([]string{"title", "grupo", "description", "priority", "data_type"}),
				},
			).Table("_system_properties").Create(&records)

		if rs.Error != nil {
			return errs.Pgf(rs.Error)
		}
	}
	r.ready.Store(empresa, true)
	return nil
}

func (r *SystemPropsPgProvider) GetStr(ctx context.Context, key properties.SystemProperty) (string, error) {
	if err := r.ensureTable(ctx, identitysdk.Empresa(ctx)); err != nil {
		return "", err
	}
	keyStr := identitysdk.Empresa(ctx, string(key))
	conn := r.manager.Conn(ctx)
	var item PropItem
	err := conn.Where("key = ?", keyStr).Select("key", "value").Find(&item).Error
	if err != nil {
		return "", errs.Pgf(err)
	}
	if item.Key != keyStr {
		return "", properties.NewPropertyNotFoundError(keyStr)
	}
	return item.Value, nil
}

func (r *SystemPropsPgProvider) GetBool(ctx context.Context, key properties.SystemProperty) (bool, error) {
	strVal, err := r.GetStr(ctx, key)
	if err != nil {
		return false, err
	}

	switch strings.ToLower(strings.TrimSpace(strVal)) {
	case "1", "t", "true", "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func (r *SystemPropsPgProvider) GetDecimal(ctx context.Context, key properties.SystemProperty) (decimal.Decimal, error) {
	strVal, err := r.GetStr(ctx, key)
	if err != nil {
		return decimal.Zero, err
	}
	var _float64 = cast.ToFloat64(strVal)
	return decimal.NewFromFloat(_float64), nil
}

func (r *SystemPropsPgProvider) GetInt(ctx context.Context, key properties.SystemProperty) (int64, error) {
	_dec, err := r.GetDecimal(ctx, key)
	if err != nil {
		return 0, err
	}
	return _dec.IntPart(), nil
}

func (r *SystemPropsPgProvider) GetJSON(ctx context.Context, key properties.SystemProperty, target any) error {
	strVal, err := r.GetStr(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(strVal), target)
}

// RetriveAll implements properties.SystemPropertiesMutator.
func (r *SystemPropsPgProvider) RetriveAll(ctx context.Context) ([]models.DetailedSystemProperty, error) {
	if err := r.ensureTable(ctx, identitysdk.Empresa(ctx)); err != nil {
		return nil, err
	}
	var prefix = identitysdk.EmpresaPrefix(ctx)
	const qry = `
		SELECT 
		key AS id, 
		title, 
		grupo, 
		description, 
		data_type AS type, 
		value
	FROM 
		_system_properties
	WHERE 
		key LIKE ?
	ORDER BY 
		priority`
	var tx = r.manager.Conn(ctx)
	var entries = []models.DetailedSystemProperty{}
	if rs := tx.Raw(qry, prefix).Scan(&entries); rs.Error != nil {
		return nil, errs.Pgf(rs.Error)
	}
	return entries, nil
}

// Update implements properties.SystemPropertiesMutator.
func (r *SystemPropsPgProvider) Update(ctx context.Context, entries []models.BasicSystemProperty) error {
	if len(entries) == 0 {
		return nil
	}
	if err := r.ensureTable(ctx, identitysdk.Empresa(ctx)); err != nil {
		return err
	}
	const qry = `
    UPDATE _system_properties AS sp
    SET value = data.value
    FROM (VALUES %s) AS data(id, value)
    WHERE sp.key = data.id`

	return r.manager.WithTx(ctx, func(ctx context.Context) error {
		tx := r.manager.Conn(ctx)

		placeholders := make([]string, 0, len(entries))
		values := make([]interface{}, 0, len(entries)*2)
		for _, entry := range entries {
			placeholders = append(placeholders, "(?, ?)")
			values = append(values, identitysdk.Empresa(ctx, entry.ID), entry.Value)
		}

		finalQry := fmt.Sprintf(qry, strings.Join(placeholders, ","))
		if err := tx.Exec(finalQry, values...).Error; err != nil {
			return errs.Pgf(err)
		}

		return nil
	})

}
