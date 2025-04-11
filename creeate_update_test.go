package identitysdk_test

import (
	"context"
	"testing"
	"time"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/entities"
)

func createContext() context.Context {
	return identitysdk.BuildContext(context.Background(), "", &entities.JwtData{
		Jwt: entities.Jwt{Username: "kevin"},
	})
}

type testStruct struct {
	CreatedBy string
	CreatedAt time.Time
	UpdatedBy string
	UpdatedAt time.Time
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
