package list

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// ResolveByValue returns the list value whose primitive value or searchable
// struct field exactly matches search after text normalization.
// Searchable fields are checked in identifier-first order.
func ResolveByValue[T comparable](list []T, search string) *T {
	search = normalizeSearchValue(search)
	if search == "" {
		return nil
	}

	var match *T
	bestPriority := len(resolveSearchFields) + 1

	for i := range list {
		value := reflect.ValueOf(list[i])
		priority, ok := resolveValuePriority(value, search)
		if ok && priority < bestPriority {
			match = &list[i]
			bestPriority = priority
		}
	}

	return match
}

var resolveSearchFields = []string{
	"codigo",
	"code",
	"id",
	"uuid",
	"key",
	"slug",
	"descripcion",
	"description",
	"name",
	"nombre",
	"label",
	"title",
	"value",
	"display_name",
	"full_name",
}

func resolveValuePriority(value reflect.Value, search string) (int, bool) {
	value = indirectValue(value)
	if !value.IsValid() {
		return 0, false
	}

	if value.Kind() != reflect.Struct {
		return len(resolveSearchFields), normalizeSearchValue(fmt.Sprint(value.Interface())) == search
	}

	for priority, fieldName := range resolveSearchFields {
		field, ok := resolveStructField(value, fieldName)
		if !ok {
			continue
		}
		field = indirectValue(field)
		if field.IsValid() && normalizeSearchValue(fmt.Sprint(field.Interface())) == search {
			return priority, true
		}
	}

	return 0, false
}

func resolveStructField(value reflect.Value, name string) (reflect.Value, bool) {
	typeInfo := value.Type()
	for i := 0; i < value.NumField(); i++ {
		fieldInfo := typeInfo.Field(i)
		jsonName, _, _ := strings.Cut(fieldInfo.Tag.Get("json"), ",")
		if strings.EqualFold(fieldInfo.Name, name) || strings.EqualFold(jsonName, name) {
			field := value.Field(i)
			if field.CanInterface() {
				return field, true
			}
		}
	}

	return reflect.Value{}, false
}

func indirectValue(value reflect.Value) reflect.Value {
	for value.IsValid() && (value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface) {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}

	return value
}

func normalizeSearchValue(value string) string {
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		r = unicode.ToLower(r)
		switch r {
		case 'á', 'à', 'ä', 'â':
			return 'a'
		case 'é', 'è', 'ë', 'ê':
			return 'e'
		case 'í', 'ì', 'ï', 'î':
			return 'i'
		case 'ó', 'ò', 'ö', 'ô':
			return 'o'
		case 'ú', 'ù', 'ü', 'û':
			return 'u'
		case 'ñ':
			return 'n'
		default:
			return r
		}
	}, value))
}
