package properties

import (
	"context"

	"github.com/sfperusacdev/identitysdk/helpers/properties/models"
	"github.com/shopspring/decimal"
)

type SystemProperty string

type SystemPropertiesMutator interface {
	RetriveAll(ctx context.Context) ([]models.DetailedSystemProperty, error)
	Update(ctx context.Context, entries []models.BasicSystemProperty) error
}

type SystemPropsProvider interface {
	GetStr(ctx context.Context, key SystemProperty) (string, error)
	GetBool(ctx context.Context, key SystemProperty) (bool, error)
	GetDecimal(ctx context.Context, key SystemProperty) (decimal.Decimal, error)
	GetInt(ctx context.Context, key SystemProperty) (int64, error)
	GetJSON(ctx context.Context, key SystemProperty, target any) error
}
