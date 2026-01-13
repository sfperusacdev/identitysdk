package identitysdk_test

import (
	"context"
	"testing"
	"time"

	"maps"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/entities"
)

func createContext() context.Context {
	return identitysdk.BuildContextWithSucursal(context.Background(), "", "sf001", &entities.JwtData{
		Jwt:     entities.Jwt{Username: "kevin", Empresa: "sfperu"},
		Session: entities.Session{Company: "sfperu"},
	})
}

type testStruct struct {
	identitysdk.Model
}

func TestCreateBy(t *testing.T) {
	ctx := createContext()
	obj := &testStruct{}

	identitysdk.CreateBy(ctx, obj)

	if obj.CreatedBy != "kevin" {
		t.Errorf("expected CreatedBy to be 'kevin', got '%s'", obj.CreatedBy)
	}
	if obj.UpdatedBy != "kevin" {
		t.Errorf("expected UpdatedBy to be 'kevin', got '%s'", obj.UpdatedBy)
	}
	if time.Since(obj.CreatedAt) > time.Second {
		t.Errorf("CreatedAt not set properly")
	}
	if time.Since(obj.UpdatedAt) > time.Second {
		t.Errorf("UpdatedAt not set properly")
	}
}

func TestCreateByMap(t *testing.T) {
	ctx := createContext()
	obj := map[string]any{}

	identitysdk.CreateBy(ctx, &obj)

	if obj["created_by"] != "kevin" {
		t.Errorf("expected created_by to be 'kevin', got '%v'", obj["created_by"])
	}
	if obj["updated_by"] != "kevin" {
		t.Errorf("expected updated_by to be 'kevin', got '%v'", obj["updated_by"])
	}

	createdAt, ok1 := obj["created_at"].(time.Time)
	updatedAt, ok2 := obj["updated_at"].(time.Time)

	if !ok1 || time.Since(createdAt) > time.Second {
		t.Errorf("created_at not set properly")
	}
	if !ok2 || time.Since(updatedAt) > time.Second {
		t.Errorf("updated_at not set properly")
	}
}

func TestCreateBy_NonPointer(t *testing.T) {
	obj := map[string]any{
		"created_by": "initial",
		"created_at": time.Now(),
		"updated_by": "initial",
		"updated_at": time.Now(),
	}

	original := make(map[string]any)
	maps.Copy(original, obj)

	identitysdk.CreateBy(context.Background(), obj)

	for _, key := range []string{"created_by", "created_at", "updated_by", "updated_at"} {
		if obj[key] != original[key] {
			t.Errorf("expected '%s' to remain unchanged on non-pointer input", key)
		}
	}
}

func TestCreateByMap_OverwriteExistingKeys(t *testing.T) {
	ctx := createContext()
	oldTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	obj := map[string]any{
		"created_by": "someone_else",
		"created_at": oldTime,
		"updated_by": "someone_else",
		"updated_at": oldTime,
	}

	identitysdk.CreateBy(ctx, &obj)

	if obj["created_by"] != "kevin" {
		t.Errorf("expected created_by to be 'kevin', got '%v'", obj["created_by"])
	}
	if obj["updated_by"] != "kevin" {
		t.Errorf("expected updated_by to be 'kevin', got '%v'", obj["updated_by"])
	}

	createdAt, ok1 := obj["created_at"].(time.Time)
	updatedAt, ok2 := obj["updated_at"].(time.Time)

	if !ok1 || createdAt.Equal(oldTime) {
		t.Errorf("created_at was not overwritten")
	}
	if !ok2 || updatedAt.Equal(oldTime) {
		t.Errorf("updated_at was not overwritten")
	}
}

func TestUpdateBy_NonPointer(t *testing.T) {
	obj := map[string]any{
		"updated_by": "initial",
		"updated_at": time.Now(),
	}

	original := make(map[string]any)
	maps.Copy(original, obj)

	identitysdk.UpdateBy(context.Background(), obj)

	for _, key := range []string{"updated_by", "updated_at"} {
		if obj[key] != original[key] {
			t.Errorf("expected '%s' to remain unchanged on non-pointer input", key)
		}
	}
}

func TestUpdateBy(t *testing.T) {
	ctx := createContext()
	obj := &testStruct{}

	identitysdk.UpdateBy(ctx, obj)

	if obj.UpdatedBy != "kevin" {
		t.Errorf("expected UpdatedBy to be 'kevin', got '%s'", obj.UpdatedBy)
	}
	if time.Since(obj.UpdatedAt) > time.Second {
		t.Errorf("UpdatedAt not set properly")
	}
}

func TestUpdateByMap_OverwriteExistingKeys(t *testing.T) {
	ctx := createContext()
	oldTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	obj := map[string]any{
		"updated_by": "someone_else",
		"updated_at": oldTime,
	}

	identitysdk.UpdateBy(ctx, &obj)

	if obj["updated_by"] != "kevin" {
		t.Errorf("expected updated_by to be 'kevin', got '%v'", obj["updated_by"])
	}

	updatedAt, ok := obj["updated_at"].(time.Time)
	if !ok || updatedAt.Equal(oldTime) {
		t.Errorf("updated_at was not overwritten")
	}
}

func TestUpdateByMap(t *testing.T) {
	ctx := createContext()
	obj := map[string]any{
		"created_by": "alice",
		"created_at": time.Now().Add(-time.Hour),
	}

	identitysdk.UpdateBy(ctx, &obj)

	if obj["created_by"] != "alice" {
		t.Errorf("expected created_by to remain 'alice', got '%v'", obj["created_by"])
	}
	if obj["updated_by"] != "kevin" {
		t.Errorf("expected updated_by to be 'kevin', got '%v'", obj["updated_by"])
	}

	updatedAt, ok := obj["updated_at"].(time.Time)
	if !ok || time.Since(updatedAt) > time.Second {
		t.Errorf("updated_at not set properly")
	}
}

type auditableStruct struct {
	CreatedBy *string
	CreatedAt *time.Time
	UpdatedBy *string
	UpdatedAt *time.Time
}

func TestCreateByWithPointers(t *testing.T) {
	ctx := createContext()
	obj := &auditableStruct{}

	identitysdk.CreateBy(ctx, obj)

	if obj.CreatedBy == nil || *obj.CreatedBy != "kevin" {
		t.Errorf("expected CreatedBy to be 'kevin', got '%v'", obj.CreatedBy)
	}
	if obj.UpdatedBy == nil || *obj.UpdatedBy != "kevin" {
		t.Errorf("expected UpdatedBy to be 'kevin', got '%v'", obj.UpdatedBy)
	}
	if obj.CreatedAt == nil || time.Since(*obj.CreatedAt) > time.Second {
		t.Errorf("CreatedAt not set properly")
	}
	if obj.UpdatedAt == nil || time.Since(*obj.UpdatedAt) > time.Second {
		t.Errorf("UpdatedAt not set properly")
	}
}

func TestUpdateByWithPointers(t *testing.T) {
	ctx := createContext()
	obj := &auditableStruct{}

	identitysdk.UpdateBy(ctx, obj)

	if obj.UpdatedBy == nil || *obj.UpdatedBy != "kevin" {
		t.Errorf("expected UpdatedBy to be 'kevin', got '%v'", obj.UpdatedBy)
	}
	if obj.UpdatedAt == nil || time.Since(*obj.UpdatedAt) > time.Second {
		t.Errorf("UpdatedAt not set properly")
	}
}
