package identitysdk

import (
	"context"
	"time"

	"log/slog"

	"github.com/jinzhu/copier"
)

type update struct {
	UpdatedBy string
	UpdatedAt time.Time
}

type create struct {
	CreatedBy string
	CreatedAt time.Time
}

func CreateBy(ctx context.Context, i any) {
	createdBy := create{
		CreatedBy: Username(ctx),
		CreatedAt: time.Now(),
	}
	if err := copier.Copy(i, &createdBy); err != nil {
		slog.Warn("CreateBy copy error", "error", err)
	}
}

func UpdateBy(ctx context.Context, i any) {
	updatedBy := update{
		UpdatedBy: Username(ctx),
		UpdatedAt: time.Now(),
	}
	if err := copier.Copy(i, &updatedBy); err != nil {
		slog.Warn("UpdateBy copy error", "error", err)
	}
}
