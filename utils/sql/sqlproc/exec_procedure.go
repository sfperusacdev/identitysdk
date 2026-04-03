package sqlproc

import (
	"errors"
	"strings"
	"unicode"

	"github.com/sfperusacdev/identitysdk/utils/sql/sqlsanitize"
)

var ErrStoredProcedureNotFound = errors.New("stored procedure identifier not found")

func GetStoredProcedureIdentifierFromQuery(query string) (string, error) {
	cleaned := sqlsanitize.RemoveComments(query)

	start, end := findStoredProcedureIdentifierBounds(cleaned)
	if start == -1 {
		return "", ErrStoredProcedureNotFound
	}

	return normalizeStoredProcedureIdentifier(cleaned[start:end]), nil
}

func ReplaceStoredProcedureIdentifierInQuery(query string, newIdentifier string) (string, error) {
	cleaned := sqlsanitize.RemoveComments(query)

	start, end := findStoredProcedureIdentifierBounds(cleaned)
	if start == -1 {
		return "", ErrStoredProcedureNotFound
	}

	return cleaned[:start] + newIdentifier + cleaned[end:], nil
}

func findStoredProcedureIdentifierBounds(query string) (int, int) {
	for i := 0; i < len(query); {
		i = skipWhitespace(query, i)
		if i >= len(query) {
			return -1, -1
		}

		wordStart := i

		for i < len(query) && isIdentifierChar(query[i]) {
			i++
		}

		if wordStart == i {
			i++
			continue
		}

		word := query[wordStart:i]
		if !equalsFoldASCII(word, "exec") && !equalsFoldASCII(word, "execute") {
			continue
		}

		i = skipWhitespace(query, i)
		if i >= len(query) {
			return -1, -1
		}

		identifierStart := i
		identifierEnd := consumeStoredProcedureIdentifier(query, i)
		if identifierEnd == identifierStart {
			return -1, -1
		}

		return identifierStart, identifierEnd
	}

	return -1, -1
}

func consumeStoredProcedureIdentifier(query string, start int) int {
	pos := start
	foundPart := false

	for {
		pos = skipWhitespace(query, pos)

		partEnd := consumeIdentifierPart(query, pos)
		if partEnd == pos {
			break
		}

		foundPart = true
		pos = partEnd

		next := skipWhitespace(query, pos)
		if next < len(query) && query[next] == '.' {
			pos = next + 1
			continue
		}

		return partEnd
	}

	if !foundPart {
		return start
	}

	return pos
}

func consumeIdentifierPart(query string, start int) int {
	if start >= len(query) {
		return start
	}

	if query[start] == '[' {
		i := start + 1

		for i < len(query) && query[i] != ']' {
			i++
		}

		if i < len(query) && query[i] == ']' {
			return i + 1
		}

		return start
	}

	i := start

	for i < len(query) && isIdentifierChar(query[i]) {
		i++
	}

	return i
}

func normalizeStoredProcedureIdentifier(identifier string) string {
	parts := splitIdentifierParts(identifier)
	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, ".")
}

func splitIdentifierParts(identifier string) []string {
	var parts []string

	for i := 0; i < len(identifier); {
		i = skipWhitespace(identifier, i)

		if i >= len(identifier) {
			break
		}

		if identifier[i] == '.' {
			i++
			continue
		}

		if identifier[i] == '[' {
			j := i + 1

			for j < len(identifier) && identifier[j] != ']' {
				j++
			}

			if j >= len(identifier) {
				break
			}

			parts = append(parts, identifier[i:j+1])
			i = j + 1
			continue
		}

		j := i

		for j < len(identifier) && isIdentifierChar(identifier[j]) {
			j++
		}

		if j == i {
			break
		}

		parts = append(parts, identifier[i:j])
		i = j
	}

	return parts
}

func skipWhitespace(query string, start int) int {
	i := start

	for i < len(query) && unicode.IsSpace(rune(query[i])) {
		i++
	}

	return i
}

func isIdentifierChar(b byte) bool {
	return b == '_' ||
		(b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9')
}

func equalsFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		aa := a[i]
		bb := b[i]

		if aa >= 'A' && aa <= 'Z' {
			aa += 'a' - 'A'
		}

		if bb >= 'A' && bb <= 'Z' {
			bb += 'a' - 'A'
		}

		if aa != bb {
			return false
		}
	}

	return true
}
