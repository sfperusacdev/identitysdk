package sqlutil

import (
	"fmt"
	"strings"
)

func NormalizeSQLServerIdentifier(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("identifier is empty")
	}

	parts, err := splitSQLServerIdentifier(name)
	if err != nil {
		return "", err
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("identifier is empty")
	}

	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return "", fmt.Errorf("invalid identifier: %q", name)
		}
		normalized = append(normalized, "["+escapeSQLServerIdentifierPart(unquoteBracketedIdentifier(part))+"]")
	}

	return strings.Join(normalized, "."), nil
}

func splitSQLServerIdentifier(name string) ([]string, error) {
	var parts []string
	var current strings.Builder
	inBracket := false

	for i := 0; i < len(name); i++ {
		ch := name[i]

		switch ch {
		case '[':
			if inBracket {
				current.WriteByte(ch)
				continue
			}
			inBracket = true
			current.WriteByte(ch)

		case ']':
			if !inBracket {
				return nil, fmt.Errorf("invalid identifier: %q", name)
			}
			if i+1 < len(name) && name[i+1] == ']' {
				current.WriteString("]]")
				i++
				continue
			}
			inBracket = false
			current.WriteByte(ch)

		case '.':
			if inBracket {
				current.WriteByte(ch)
				continue
			}
			part := strings.TrimSpace(current.String())
			if part == "" {
				return nil, fmt.Errorf("invalid identifier: %q", name)
			}
			parts = append(parts, part)
			current.Reset()

		default:
			current.WriteByte(ch)
		}
	}

	if inBracket {
		return nil, fmt.Errorf("unterminated bracketed identifier: %q", name)
	}

	last := strings.TrimSpace(current.String())
	if last == "" {
		return nil, fmt.Errorf("invalid identifier: %q", name)
	}

	parts = append(parts, last)
	return parts, nil
}

func unquoteBracketedIdentifier(part string) string {
	part = strings.TrimSpace(part)
	if len(part) >= 2 && part[0] == '[' && part[len(part)-1] == ']' {
		part = part[1 : len(part)-1]
	}
	return strings.ReplaceAll(part, "]]", "]")
}

func escapeSQLServerIdentifierPart(part string) string {
	return strings.ReplaceAll(part, "]", "]]")
}
