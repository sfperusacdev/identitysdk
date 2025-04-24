package identitysdk

import (
	"context"
	"reflect"
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
	var typ = reflect.TypeOf(i)
	if typ.Kind() != reflect.Pointer {
		slog.Warn("CreateBy expects a pointer to a map or struct", "got", typ.Kind())
		return
	}
	var username = Username(ctx)
	var now = time.Now()

	typ = typ.Elem()
	if typ.Kind() == reflect.Map {
		var v = reflect.ValueOf(i).Elem()
		var vnow = reflect.ValueOf(now)
		var vusername = reflect.ValueOf(username)
		v.SetMapIndex(reflect.ValueOf("created_at"), vnow)
		v.SetMapIndex(reflect.ValueOf("created_by"), vusername)
		v.SetMapIndex(reflect.ValueOf("updated_at"), vnow)
		v.SetMapIndex(reflect.ValueOf("updated_by"), vusername)
		return
	}

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
	var typ = reflect.TypeOf(i)
	if typ.Kind() != reflect.Pointer {
		slog.Warn("UpdateBy expects a pointer to a map or struct", "got", typ.Kind())
		return
	}
	var username = Username(ctx)
	var now = time.Now()

	typ = typ.Elem()
	if typ.Kind() == reflect.Map {
		var v = reflect.ValueOf(i).Elem()
		var vnow = reflect.ValueOf(now)
		var vusername = reflect.ValueOf(username)
		v.SetMapIndex(reflect.ValueOf("updated_at"), vnow)
		v.SetMapIndex(reflect.ValueOf("updated_by"), vusername)
		return
	}

	updatedBy := update{
		UpdatedBy: username,
		UpdatedAt: now,
	}
	if err := copier.Copy(i, &updatedBy); err != nil {
		slog.Warn("UpdateBy copy error", "error", err)
	}
}
