package sqlproc

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
)

func RenameProcedureDefinition(input string) (string, string, error) {
	newName := "p_" + strings.ReplaceAll(uuid.NewString(), "-", "")

	newDef, err := ReplaceProcedureName(input, newName)
	if err != nil {
		return "", "", err
	}

	return newDef, newName, nil
}

func ReplaceProcedureName(input string, newName string) (string, error) {
	source := strings.TrimSpace(input)
	if source == "" {
		return "", fmt.Errorf("input is empty")
	}

	if strings.TrimSpace(newName) == "" {
		return "", fmt.Errorf("new procedure name is empty")
	}

	if err := ValidateProcedureDefinition(source); err != nil {
		slog.Error("procedure definition validation failed", "error", err)
		return "", fmt.Errorf("invalid procedure definition")
	}

	nameStart, nameEnd, err := locateProcedureNameRange(source)
	if err != nil {
		return "", err
	}

	return source[:nameStart] + newName + source[nameEnd:], nil
}

func locateProcedureNameRange(source string) (int, int, error) {
	sc := newScanner(source)

	if err := sc.readCreateClause(); err != nil {
		return 0, 0, err
	}

	if err := sc.readProcedureKeyword(); err != nil {
		return 0, 0, err
	}

	sc.skipWhitespaceAndComments()
	nameStart := sc.pos

	if _, err := sc.readNamePart(); err != nil {
		return 0, 0, fmt.Errorf("expected procedure name at position %d", nameStart)
	}

	checkpoint := sc.pos
	sc.skipWhitespaceAndComments()

	if sc.peek() == '.' {
		sc.pos++
		sc.skipWhitespaceAndComments()

		if _, err := sc.readNamePart(); err != nil {
			return 0, 0, fmt.Errorf("invalid procedure name after schema qualifier at position %d", sc.pos)
		}
	} else {
		sc.pos = checkpoint
	}

	return nameStart, sc.pos, nil
}
