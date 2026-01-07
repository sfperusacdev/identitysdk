package scripting

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/dop251/goja"
)

type ScriptCommonService struct {
	VarName string
}

func NewScriptCommonService() *ScriptCommonService {
	return &ScriptCommonService{VarName: "ctx"}
}

func (s *ScriptCommonService) Execute(ptr any, script string) error {
	b, err := json.Marshal(ptr)
	if err != nil {
		slog.Error("marshal input", "error", err)
		return err
	}

	vm := goja.New()

	_, err = vm.RunString(fmt.Sprintf(`%s = JSON.parse(%q);`, s.VarName, b))
	if err != nil {
		slog.Error("js parse error", "error", err)
		return err
	}

	_, err = vm.RunString(script)
	if err != nil {
		slog.Error("script error", "error", err)
		return err
	}

	exported := vm.Get(s.VarName).Export()

	if err := validateTypes(ptr, exported); err != nil {
		slog.Error("type validation failed", "error", err)
		return err
	}

	outb, err := json.Marshal(exported)
	if err != nil {
		slog.Error("marshal output", "error", err)
		return err
	}

	if err := json.Unmarshal(outb, ptr); err != nil {
		slog.Error("unmarshal modified", "error", err)
		return err
	}

	return nil
}

func validateTypes(ptr any, data any) error {
	return validateValue(reflect.ValueOf(ptr).Elem(), data)
}
func validateValue(dst reflect.Value, src any) error {
	t := dst.Type()

	if t == reflect.TypeFor[time.Time]() {
		if _, ok := src.(string); ok {
			return nil
		}
		return fmt.Errorf("expected time.Time ISO string, got %T", src)
	}

	switch dst.Kind() {

	case reflect.Struct:
		m, ok := src.(map[string]any)
		if !ok {
			return fmt.Errorf("expected struct, got %T", src)
		}
		for i := 0; i < dst.NumField(); i++ {
			f := dst.Type().Field(i)
			if !dst.Field(i).CanSet() {
				continue
			}
			srcVal, exists := m[f.Name]
			if !exists {
				continue
			}
			if err := validateValue(dst.Field(i), srcVal); err != nil {
				return err
			}
		}

	case reflect.Map:
		m, ok := src.(map[string]any)
		if !ok {
			return fmt.Errorf("expected map, got %T", src)
		}
		elemType := dst.Type().Elem()
		for _, v := range m {
			tmp := reflect.New(elemType).Elem()
			if err := validateValue(tmp, v); err != nil {
				return err
			}
		}

	case reflect.Slice:
		arr, ok := src.([]any)
		if !ok {
			return fmt.Errorf("expected slice, got %T", src)
		}
		elemType := dst.Type().Elem()
		for _, v := range arr {
			tmp := reflect.New(elemType).Elem()
			if err := validateValue(tmp, v); err != nil {
				return err
			}
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch src.(type) {
		case float64, int, int64:
			return nil
		default:
			return fmt.Errorf("expected int-type, got %T", src)
		}

	case reflect.Float32, reflect.Float64:
		switch src.(type) {
		case float64, int, int64:
			return nil
		default:
			return fmt.Errorf("expected float-type, got %T", src)
		}

	case reflect.String:
		if _, ok := src.(string); !ok {
			return fmt.Errorf("expected string, got %T", src)
		}

	case reflect.Bool:
		if _, ok := src.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", src)
		}

	case reflect.Interface:
		return nil

	default:
		return fmt.Errorf("unsupported type: %s", dst.Kind())
	}

	return nil
}
