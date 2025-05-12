package properties

import (
	"context"

	"github.com/shopspring/decimal"
)

type SystemPropsProvider interface {
	GetStr(ctx context.Context, key string) (string, error)
	GetBool(ctx context.Context, key string) (bool, error)
	GetDecimal(ctx context.Context, key string) (decimal.Decimal, error)
	GetInt(ctx context.Context, key string) (int64, error)
	GetJSON(ctx context.Context, key string, target any) error
}
