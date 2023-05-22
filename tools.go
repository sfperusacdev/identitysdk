package identitysdk

import (
	"context"
	"fmt"
	"time"

	"github.com/jinzhu/copier"
)

type update struct {
	UpdatedBy string
	UpdatedAt time.Time
}
type create struct {
	CreatedBy string
	CreatedAt time.Time
	UpdatedBy string
	UpdatedAt time.Time
}

func CreateBy(ctx context.Context, i any) {
	cratedBy := create{
		CreatedBy: Username(ctx),
		CreatedAt: time.Now(),
	}
	if err := copier.Copy(i, &cratedBy); err != nil {
		if logger != nil {
			logger.Warn(fmt.Sprintf("createBy copy error: %s", err.Error()))
		}
	}
}
func UpdateBy(ctx context.Context, i any) {
	updateBy := update{
		UpdatedBy: Username(ctx),
		UpdatedAt: time.Now(),
	}
	if err := copier.Copy(i, &updateBy); err != nil {
		if logger != nil {
			logger.Warn(fmt.Sprintf("createBy copy error: %s", err.Error()))
		}
	}
}
