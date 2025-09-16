package fotocheck

import (
	"strings"
	"unicode"
)

func Abbreviate(fullName string) string {
	particles := map[string]bool{
		"de": true, "del": true, "la": true, "las": true, "los": true,
	}

	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return ""
	}

	var result []string
	result = append(result, parts[0]) // nombre principal

	i := 1
	for i < len(parts) {
		word := strings.ToLower(parts[i])
		if particles[word] && i+1 < len(parts) {
			compound := parts[i] + " " + parts[i+1]
			result = append(result, string(unicode.ToUpper(rune(compound[0])))+".")
			i += 2
		} else {
			result = append(result, string(unicode.ToUpper(rune(parts[i][0])))+".")
			i++
		}
	}

	return strings.Join(result, " ")
}

func AbbreviateIfLonger(fullName string, maxLen int) string {
	if len(fullName) <= maxLen {
		return fullName
	}
	return Abbreviate(fullName)
}

func ToLower(s string) string {
	return strings.ToLower(s)
}

func ToUpper(s string) string {
	return strings.ToUpper(s)
}

func Capitalize(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}
	return string(runes)
}

func CleanWhitespace(s string) string {
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

func RemoveInvisibleChars(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsPrint(r) || r == '\n' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func NormalizeText(s string) string {
	return CleanWhitespace(RemoveInvisibleChars(s))
}
