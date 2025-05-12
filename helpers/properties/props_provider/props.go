package propsprovider

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/helpers/properties"
	connection "github.com/sfperusacdev/identitysdk/pg-connection"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
	"github.com/user0608/goones/errs"
	"gorm.io/gorm"
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
}

var _ properties.SystemPropsProvider = (*SystemPropsPgProvider)(nil)

func NewSystemPropsPgProvider(manager connection.StorageManager) *SystemPropsPgProvider {
	return &SystemPropsPgProvider{manager: manager}
}

func (r *SystemPropsPgProvider) ensureTable(ctx context.Context, empresa string) error {
	if _, exists := r.ready.Load(empresa); exists {
		return nil
	}
	const script = `
	CREATE TABLE IF NOT EXISTS system_properties (
		key TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		grupo TEXT NOT NULL,
		description TEXT,
		value JSONB NOT NULL,
		priority SMALLINT DEFAULT 0 CHECK (priority >= 0)
	)`
	var tx = r.manager.Conn(ctx)
	rs := tx.Session(&gorm.Session{Logger: logger.Discard}).Exec(script)
	if rs.Error != nil {
		return errs.Pgf(rs.Error)
	}
	r.ready.Store(empresa, true)
	return nil
}

func (r *SystemPropsPgProvider) GetStr(ctx context.Context, key string) (string, error) {
	if err := r.ensureTable(ctx, identitysdk.Empresa(ctx)); err != nil {
		return "", err
	}
	key = identitysdk.Empresa(ctx, key)
	conn := r.manager.Conn(ctx)
	var item PropItem
	err := conn.Where("key = ?", key).Select("value").First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", properties.NewPropertyNotFoundError(key)
		}
		return "", err
	}
	return item.Value, nil
}

func (r *SystemPropsPgProvider) GetBool(ctx context.Context, key string) (bool, error) {
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

func (r *SystemPropsPgProvider) GetDecimal(ctx context.Context, key string) (decimal.Decimal, error) {
	strVal, err := r.GetStr(ctx, key)
	if err != nil {
		return decimal.Zero, err
	}
	var _float64 = cast.ToFloat64(strVal)
	return decimal.NewFromFloat(_float64), nil
}

func (r *SystemPropsPgProvider) GetInt(ctx context.Context, key string) (int64, error) {
	_dec, err := r.GetDecimal(ctx, key)
	if err != nil {
		return 0, err
	}
	return _dec.IntPart(), nil
}

func (r *SystemPropsPgProvider) GetJSON(ctx context.Context, key string, target any) error {
	strVal, err := r.GetStr(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(strVal), target)
}
