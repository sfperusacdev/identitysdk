package sqlproc

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/sfperusacdev/identitysdk/utils/randomtext"
	"github.com/sfperusacdev/identitysdk/utils/sql/sqlutil"
)

type ProcedureDefinition struct {
	Name          string
	SqlDefinition string
}

func RenameProcedureWithRandomName(input string) (ProcedureDefinition, error) {
	return renameProcedureWithRandomName(input, false)
}

func RenameProcedureWithTemporalRandomName(input string) (ProcedureDefinition, error) {
	return renameProcedureWithRandomName(input, true)
}

func renameProcedureWithRandomName(input string, temporal bool) (ProcedureDefinition, error) {
	source := strings.TrimSpace(input)
	if source == "" {
		return ProcedureDefinition{}, fmt.Errorf("input is empty")
	}

	originalName, err := ExtractProcedureName(source)
	if err != nil {
		return ProcedureDefinition{}, err
	}
	parsed, err := sqlutil.ParseSQLServerIdentifier(originalName)
	if err != nil {
		return ProcedureDefinition{}, err
	}
	name := fmt.Sprintf("p_%s", randomtext.String(20))
	if temporal {
		name = fmt.Sprintf("#%s", name)
	}
	procedureIdentifier := sqlutil.SQLServerIdentifier{
		ObjectName: fmt.Sprintf("p_%s", randomtext.String(20)),
		SchemaPath: parsed.SchemaPath,
	}

	return RenameProcedure(source, procedureIdentifier.String())
}

func RenameProcedure(input string, newName string) (ProcedureDefinition, error) {
	source := strings.TrimSpace(input)
	if source == "" {
		return ProcedureDefinition{}, fmt.Errorf("input is empty")
	}

	if strings.TrimSpace(newName) == "" {
		return ProcedureDefinition{}, fmt.Errorf("new procedure name is empty")
	}

	if err := ValidateProcedureDefinition(source); err != nil {
		slog.Error("procedure definition validation failed", "error", err)
		return ProcedureDefinition{}, fmt.Errorf("invalid procedure definition")
	}

	nameStart, nameEnd, err := locateProcedureNameRange(source)
	if err != nil {
		return ProcedureDefinition{}, err
	}

	definition := source[:nameStart] + newName + source[nameEnd:]

	return ProcedureDefinition{
		Name:          newName,
		SqlDefinition: definition,
	}, nil
}

func ExtractProcedureName(input string) (string, error) {
	source := strings.TrimSpace(input)
	if source == "" {
		return "", fmt.Errorf("input is empty")
	}

	if err := ValidateProcedureDefinition(source); err != nil {
		return "", fmt.Errorf("invalid procedure definition")
	}

	nameStart, nameEnd, err := locateProcedureNameRange(source)
	if err != nil {
		return "", err
	}

	return source[nameStart:nameEnd], nil
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
