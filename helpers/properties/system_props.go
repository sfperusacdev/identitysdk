package properties

import (
	"context"

	"github.com/sfperusacdev/identitysdk/helpers/properties/models"
	"github.com/shopspring/decimal"
)

type SystemPropertiesMutator interface {
	RetriveAll(ctx context.Context) ([]models.DetailedSystemProperty, error)
	Update(ctx context.Context, entries []models.BasicSystemProperty) error
}

type SystemPropsProvider interface {
	GetStr(ctx context.Context, key string) (string, error)
	GetBool(ctx context.Context, key string) (bool, error)
	GetDecimal(ctx context.Context, key string) (decimal.Decimal, error)
	GetInt(ctx context.Context, key string) (int64, error)
	GetJSON(ctx context.Context, key string, target any) error
}
