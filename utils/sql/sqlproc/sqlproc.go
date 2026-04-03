package sqlproc

import (
	"fmt"
	"strings"
	"unicode"
)

func ValidateProcedureDefinition(input string) error {
	source := strings.TrimSpace(input)
	if source == "" {
		return newValidationError("input is empty", source, 0)
	}

	sc := newScanner(source)

	if err := sc.readProcedureHeader(); err != nil {
		return err
	}

	if err := sc.readRequiredKeyword("AS", "expected AS after procedure header"); err != nil {
		return err
	}

	if err := sc.validateProcedureBodyStructure(); err != nil {
		return err
	}

	return nil
}

type ValidationError struct {
	Message  string
	Position int
	Line     int
	Column   int
	Snippet  string
	Pointer  string
}

func (e *ValidationError) Error() string {
	if e.Snippet == "" {
		return fmt.Sprintf("%s at line %d, column %d", e.Message, e.Line, e.Column)
	}

	return fmt.Sprintf(
		"%s at line %d, column %d\n%s\n%s",
		e.Message,
		e.Line,
		e.Column,
		e.Snippet,
		e.Pointer,
	)
}

func newValidationError(message string, source string, position int) error {
	line, column, snippet, pointer := buildErrorContext(source, position)
	return &ValidationError{
		Message:  message,
		Position: position,
		Line:     line,
		Column:   column,
		Snippet:  snippet,
		Pointer:  pointer,
	}
}

func buildErrorContext(source string, position int) (int, int, string, string) {
	if position < 0 {
		position = 0
	}
	if position > len(source) {
		position = len(source)
	}

	line := 1
	column := 1
	lineStart := 0

	for i := 0; i < position; i++ {
		if source[i] == '\n' {
			line++
			column = 1
			lineStart = i + 1
		} else {
			column++
		}
	}

	lineEnd := len(source)
	for i := lineStart; i < len(source); i++ {
		if source[i] == '\n' {
			lineEnd = i
			break
		}
	}

	snippet := source[lineStart:lineEnd]
	if snippet == "" {
		return line, column, snippet, ""
	}

	pointerPos := position - lineStart
	if pointerPos < 0 {
		pointerPos = 0
	}
	if pointerPos > len(snippet) {
		pointerPos = len(snippet)
	}

	pointer := strings.Repeat(" ", pointerPos) + "^"
	return line, column, snippet, pointer
}

var errNotAStringLiteral = fmt.Errorf("not a string literal")

type scanner struct {
	source string
	pos    int
	length int
}

func newScanner(source string) *scanner {
	return &scanner{
		source: source,
		pos:    0,
		length: len(source),
	}
}

func (s *scanner) eof() bool {
	return s.pos >= s.length
}

func (s *scanner) peek() byte {
	if s.eof() {
		return 0
	}
	return s.source[s.pos]
}

func (s *scanner) skipWhitespace() {
	for !s.eof() && unicode.IsSpace(rune(s.source[s.pos])) {
		s.pos++
	}
}

func (s *scanner) skipLineComment() bool {
	if s.pos+1 >= s.length {
		return false
	}

	if s.source[s.pos] == '-' && s.source[s.pos+1] == '-' {
		s.pos += 2
		for !s.eof() && s.source[s.pos] != '\n' {
			s.pos++
		}
		return true
	}

	return false
}

func (s *scanner) skipBlockComment() bool {
	if s.pos+1 >= s.length {
		return false
	}

	if s.source[s.pos] == '/' && s.source[s.pos+1] == '*' {
		start := s.pos
		s.pos += 2

		for s.pos+1 < s.length {
			if s.source[s.pos] == '*' && s.source[s.pos+1] == '/' {
				s.pos += 2
				return true
			}
			s.pos++
		}

		s.pos = s.length
		_ = start
		return true
	}

	return false
}

func (s *scanner) skipWhitespaceAndComments() bool {
	start := s.pos

	for {
		s.skipWhitespace()

		if s.skipLineComment() {
			continue
		}

		if s.skipBlockComment() {
			continue
		}

		break
	}

	return s.pos > start
}

func (s *scanner) tryReadStringLiteral() error {
	if s.peek() != '\'' {
		return errNotAStringLiteral
	}

	start := s.pos
	s.pos++

	for !s.eof() {
		if s.source[s.pos] == '\'' {
			if s.pos+1 < s.length && s.source[s.pos+1] == '\'' {
				s.pos += 2
				continue
			}
			s.pos++
			return nil
		}
		s.pos++
	}

	return newValidationError("unterminated string literal", s.source, start)
}

func (s *scanner) readToken() (string, bool) {
	if s.eof() {
		return "", false
	}

	start := s.pos

	if s.source[s.pos] == '[' {
		s.pos++
		for !s.eof() && s.source[s.pos] != ']' {
			s.pos++
		}
		if s.eof() {
			return "", false
		}
		s.pos++
		return s.source[start:s.pos], true
	}

	r := rune(s.source[s.pos])
	if !unicode.IsLetter(r) && s.source[s.pos] != '_' && s.source[s.pos] != '@' && s.source[s.pos] != '#' {
		return "", false
	}

	s.pos++

	for !s.eof() {
		r = rune(s.source[s.pos])
		if !unicode.IsLetter(r) &&
			!unicode.IsDigit(r) &&
			s.source[s.pos] != '_' &&
			s.source[s.pos] != '@' &&
			s.source[s.pos] != '#' {
			break
		}
		s.pos++
	}

	return s.source[start:s.pos], true
}

func (s *scanner) readRequiredKeyword(expected string, message string) error {
	s.skipWhitespaceAndComments()

	start := s.pos
	token, ok := s.readToken()
	if !ok {
		return newValidationError(message, s.source, start)
	}

	if !strings.EqualFold(token, expected) {
		return newValidationError(message, s.source, start)
	}

	return nil
}

func (s *scanner) readProcedureHeader() error {
	s.skipWhitespaceAndComments()

	if err := s.readCreateClause(); err != nil {
		return err
	}

	if err := s.readProcedureKeyword(); err != nil {
		return err
	}

	s.skipWhitespaceAndComments()

	if err := s.readProcedureName(); err != nil {
		return err
	}

	for {
		checkpoint := s.pos

		s.skipWhitespaceAndComments()

		token, ok := s.readToken()
		if ok {
			if strings.EqualFold(token, "AS") {
				s.pos = checkpoint
				return nil
			}
			continue
		}

		if err := s.tryReadStringLiteral(); err == nil {
			continue
		} else if err != errNotAStringLiteral {
			return err
		}

		if s.skipLineComment() || s.skipBlockComment() {
			continue
		}

		if s.eof() {
			return newValidationError("missing AS in procedure definition", s.source, s.pos)
		}

		s.pos++

		if s.pos == checkpoint {
			return newValidationError("could not parse procedure header", s.source, s.pos)
		}
	}
}
func (s *scanner) validateProcedureBodyStructure() error {
	s.skipWhitespaceAndComments()

	if s.eof() {
		return newValidationError("expected procedure body after AS", s.source, s.pos)
	}

	bodyStart := s.pos

	blockDepth := 0
	caseDepth := 0
	sawBeginBlock := false

	for !s.eof() {
		if s.skipWhitespaceAndComments() {
			continue
		}

		if err := s.tryReadStringLiteral(); err == nil {
			continue
		} else if err != errNotAStringLiteral {
			return err
		}

		tokenStart := s.pos
		token, ok := s.readToken()
		if !ok {
			s.pos++
			continue
		}

		switch strings.ToUpper(token) {
		case "BEGIN":
			nextToken, _ := s.peekNextSignificantToken()
			if strings.EqualFold(nextToken, "TRAN") || strings.EqualFold(nextToken, "TRANSACTION") {
				continue
			}
			sawBeginBlock = true
			blockDepth++

		case "CASE":
			caseDepth++

		case "END":
			if caseDepth > 0 {
				caseDepth--
				continue
			}

			if blockDepth > 0 {
				blockDepth--
				continue
			}

			return newValidationError("unexpected END token", s.source, tokenStart)
		}
	}

	if caseDepth > 0 {
		return newValidationError("missing END for CASE expression", s.source, s.length)
	}

	if sawBeginBlock && blockDepth > 0 {
		return newValidationError("missing END for BEGIN block", s.source, s.length)
	}

	if bodyStart >= s.length {
		return newValidationError("expected procedure body after AS", s.source, bodyStart)
	}

	return nil
}

func (s *scanner) peekNextSignificantToken() (string, bool) {
	checkpoint := s.pos
	defer func() {
		s.pos = checkpoint
	}()

	for {
		s.skipWhitespaceAndComments()

		if err := s.tryReadStringLiteral(); err == nil {
			continue
		} else if err != errNotAStringLiteral {
			return "", false
		}

		token, ok := s.readToken()
		if ok {
			return token, true
		}

		if s.eof() {
			return "", false
		}

		s.pos++
	}
}

func (s *scanner) readCreateClause() error {
	start := s.pos

	token, ok := s.readToken()
	if !ok {
		return newValidationError("expected CREATE, ALTER, or CREATE OR ALTER", s.source, start)
	}

	if strings.EqualFold(token, "ALTER") {
		return nil
	}

	if !strings.EqualFold(token, "CREATE") {
		return newValidationError("expected CREATE, ALTER, or CREATE OR ALTER", s.source, start)
	}

	checkpoint := s.pos
	s.skipWhitespaceAndComments()

	token, ok = s.readToken()
	if !ok {
		s.pos = checkpoint
		return nil
	}

	if !strings.EqualFold(token, "OR") {
		s.pos = checkpoint
		return nil
	}

	if err := s.readRequiredKeyword("ALTER", "expected ALTER after CREATE OR"); err != nil {
		return err
	}

	return nil
}

func (s *scanner) readProcedureKeyword() error {
	s.skipWhitespaceAndComments()

	start := s.pos
	token, ok := s.readToken()
	if !ok {
		return newValidationError("expected PROCEDURE or PROC", s.source, start)
	}

	if strings.EqualFold(token, "PROCEDURE") || strings.EqualFold(token, "PROC") {
		return nil
	}

	return newValidationError("expected PROCEDURE or PROC", s.source, start)
}

func (s *scanner) readProcedureName() error {
	start := s.pos

	if _, err := s.readNamePart(); err != nil {
		return newValidationError("expected procedure name", s.source, start)
	}

	checkpoint := s.pos
	s.skipWhitespaceAndComments()

	if s.peek() != '.' {
		s.pos = checkpoint
		return nil
	}

	s.pos++
	s.skipWhitespaceAndComments()

	if _, err := s.readNamePart(); err != nil {
		return newValidationError("invalid procedure name after schema qualifier", s.source, s.pos)
	}

	return nil
}

func (s *scanner) readNamePart() (string, error) {
	if s.eof() {
		return "", newValidationError("unexpected empty identifier", s.source, s.pos)
	}

	start := s.pos

	if s.source[s.pos] == '[' {
		s.pos++
		for !s.eof() && s.source[s.pos] != ']' {
			s.pos++
		}
		if s.eof() {
			return "", newValidationError("unterminated bracketed identifier", s.source, start)
		}
		s.pos++
		return s.source[start:s.pos], nil
	}

	r := rune(s.source[s.pos])
	if !unicode.IsLetter(r) && s.source[s.pos] != '_' && s.source[s.pos] != '#' {
		return "", newValidationError("invalid identifier", s.source, start)
	}

	s.pos++

	for !s.eof() {
		r = rune(s.source[s.pos])
		if !unicode.IsLetter(r) &&
			!unicode.IsDigit(r) &&
			s.source[s.pos] != '_' &&
			s.source[s.pos] != '#' {
			break
		}
		s.pos++
	}

	return s.source[start:s.pos], nil
}
