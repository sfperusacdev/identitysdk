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
	UpdatedBy string
	UpdatedAt time.Time
	CreatedBy string
	CreatedAt time.Time
}

func CreateBy(ctx context.Context, i any) {
	var username = Username(ctx)
	var now = time.Now()
	createdBy := create{
		UpdatedBy: username,
		UpdatedAt: now,
		CreatedBy: username,
		CreatedAt: now,
	}
	if err := copier.Copy(i, &createdBy); err != nil {
		slog.Warn("CreateBy copy error", "error", err)
	}
}

func UpdateBy(ctx context.Context, i any) {
	var username = Username(ctx)
	var now = time.Now()
	updatedBy := update{
		UpdatedBy: username,
		UpdatedAt: now,
	}
	if err := copier.Copy(i, &updatedBy); err != nil {
		slog.Warn("UpdateBy copy error", "error", err)
	}
}
