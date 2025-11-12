package arraycast

import (
	"slices"
	"strings"
	"time"

	"github.com/spf13/cast"
)

func GetStringAt(arr []string, index int) string {
	if index < 0 || index >= len(arr) {
		return ""
	}
	v := strings.TrimSpace(arr[index])
	return v
}

func GetIntAt(arr []string, index int) int {
	if index < 0 || index >= len(arr) {
		return 0
	}
	v := strings.TrimSpace(arr[index])
	if v == "" {
		return 0
	}
	return cast.ToInt(v)
}

func GetFloatAt(arr []string, index int) float64 {
	if index < 0 || index >= len(arr) {
		return 0
	}
	v := strings.TrimSpace(arr[index])
	if v == "" {
		return 0
	}
	return cast.ToFloat64(v)
}

func GetBoolAt(arr []string, index int) bool {
	if index < 0 || index >= len(arr) {
		return false
	}
	v := strings.ToLower(strings.TrimSpace(arr[index]))
	if v == "" {
		return false
	}
	affirmatives := []string{
		"yes", "y", "true", "t", "1", "on", "ok", "sure", "yeah", "affirmative",
		"si", "sí", "s", "vale", "de acuerdo", "afirmativo",
	}
	return slices.Contains(affirmatives, v)
}

func GetDateAt(arr []string, index int) time.Time {
	if index < 0 || index >= len(arr) {
		return time.Time{}
	}
	v := strings.TrimSpace(arr[index])
	if v == "" {
		return time.Time{}
	}
	return cast.ToTime(v)
}
