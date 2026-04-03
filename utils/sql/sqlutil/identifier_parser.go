package sqlutil

import (
	"fmt"
	"strings"
)

type SQLServerIdentifier struct {
	ObjectName string
	SchemaPath []string
}

func (i SQLServerIdentifier) String() string {
	parts := make([]string, 0, len(i.SchemaPath)+1)

	for _, part := range i.SchemaPath {
		parts = append(parts, "["+escapeSQLServerIdentifierPart(part)+"]")
	}

	parts = append(parts, "["+escapeSQLServerIdentifierPart(i.ObjectName)+"]")

	return strings.Join(parts, ".")
}

func ParseSQLServerIdentifier(name string) (SQLServerIdentifier, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return SQLServerIdentifier{}, fmt.Errorf("identifier is empty")
	}

	rawParts, err := splitSQLServerIdentifier(name)
	if err != nil {
		return SQLServerIdentifier{}, err
	}

	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		part = strings.TrimSpace(part)
		if part == "" {
			return SQLServerIdentifier{}, fmt.Errorf("invalid identifier: %q", name)
		}
		parts = append(parts, unquoteBracketedIdentifier(part))
	}

	result := SQLServerIdentifier{
		ObjectName: parts[len(parts)-1],
	}

	if len(parts) > 1 {
		result.SchemaPath = append([]string(nil), parts[:len(parts)-1]...)
	}

	return result, nil
}
