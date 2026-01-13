package identitysdk

import (
	"context"
	"reflect"
	"time"

	"log/slog"

	"github.com/jinzhu/copier"
)

type Model struct {
	CreatedBy string
	CreatedAt time.Time
	UpdatedBy string
	UpdatedAt time.Time
}

func (m Model) IsZero() bool {
	return m.CreatedBy == "" &&
		m.CreatedAt.IsZero() &&
		m.UpdatedBy == "" &&
		m.UpdatedAt.IsZero()
}

func (m Model) IsCreated() bool {
	return m.CreatedBy != "" && !m.CreatedAt.IsZero()
}

func (m Model) IsUpdated() bool {
	return m.UpdatedBy != "" && !m.UpdatedAt.IsZero()
}

// IsStale returns true if the last update happened earlier than the given duration.
func (m Model) IsStale(d time.Duration) bool {
	if m.UpdatedAt.IsZero() {
		return false
	}
	return time.Since(m.UpdatedAt) > d
}

// Age returns the elapsed time since creation, or zero if unset.
func (m Model) Age() time.Duration {
	if m.CreatedAt.IsZero() {
		return 0
	}
	return time.Since(m.CreatedAt)
}

// UpdatedAgo returns how much time passed since the last update, or zero if unset.
func (m Model) UpdatedAgo() time.Duration {
	if m.UpdatedAt.IsZero() {
		return 0
	}
	return time.Since(m.UpdatedAt)
}

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
