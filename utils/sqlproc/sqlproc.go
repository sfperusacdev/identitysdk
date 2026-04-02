package sqlproc

import (
	"fmt"
	"strings"
	"unicode"
)

func ValidateProcedureDefinition(input string) error {
	source := strings.TrimSpace(input)
	if source == "" {
		return newValidationError("input is empty", 0)
	}

	sc := newScanner(source)

	if err := sc.readProcedureHeader(); err != nil {
		return err
	}

	if err := sc.readRequiredKeyword("AS", "expected AS after procedure header"); err != nil {
		return err
	}

	if err := sc.readRequiredKeyword("BEGIN", "expected BEGIN after AS"); err != nil {
		return err
	}

	blockDepth := 1

	for !sc.eof() {
		if sc.skipWhitespaceAndComments() {
			continue
		}

		if err := sc.tryReadStringLiteral(); err == nil {
			continue
		} else if err != errNotAStringLiteral {
			return err
		}

		token, ok := sc.readToken()
		if !ok {
			sc.pos++
			continue
		}

		switch strings.ToUpper(token) {
		case "BEGIN":
			blockDepth++
		case "END":
			blockDepth--
			if blockDepth < 0 {
				return newValidationError("unexpected END token", sc.pos)
			}

			if blockDepth == 0 {
				sc.skipWhitespaceAndComments()

				if sc.peek() == ';' {
					sc.pos++
				}

				sc.skipWhitespaceAndComments()

				if !sc.eof() {
					return newValidationError("unexpected content found after the final END of the procedure", sc.pos)
				}

				return nil
			}
		}
	}

	return newValidationError("missing END for the outermost procedure block", sc.pos)
}

type ValidationError struct {
	Message  string
	Position int
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s at position %d", e.Message, e.Position)
}

func newValidationError(message string, position int) error {
	return &ValidationError{
		Message:  message,
		Position: position,
	}
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

		returnValue := newValidationError("unterminated block comment", start)
		_ = returnValue
		s.pos = s.length
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

	return newValidationError("unterminated string literal", start)
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
		return newValidationError(message, start)
	}

	if !strings.EqualFold(token, expected) {
		return newValidationError(message, start)
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
		if ok && strings.EqualFold(token, "AS") {
			s.pos = checkpoint
			return nil
		}
		if ok {
			s.pos = checkpoint
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
			return newValidationError("missing AS in procedure definition", s.pos)
		}

		s.pos++

		if s.pos == checkpoint {
			return newValidationError("could not parse procedure header", s.pos)
		}
	}
}

func (s *scanner) readCreateClause() error {
	start := s.pos

	token, ok := s.readToken()
	if !ok {
		return newValidationError("expected CREATE, ALTER, or CREATE OR ALTER", start)
	}

	if strings.EqualFold(token, "ALTER") {
		return nil
	}

	if !strings.EqualFold(token, "CREATE") {
		return newValidationError("expected CREATE, ALTER, or CREATE OR ALTER", start)
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
		return newValidationError("expected PROCEDURE or PROC", start)
	}

	if strings.EqualFold(token, "PROCEDURE") || strings.EqualFold(token, "PROC") {
		return nil
	}

	return newValidationError("expected PROCEDURE or PROC", start)
}

func (s *scanner) readProcedureName() error {
	start := s.pos

	if _, err := s.readNamePart(); err != nil {
		return newValidationError("expected procedure name", start)
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
		return newValidationError("invalid procedure name after schema qualifier", s.pos)
	}

	return nil
}

func (s *scanner) readNamePart() (string, error) {
	if s.eof() {
		return "", newValidationError("unexpected empty identifier", s.pos)
	}

	start := s.pos

	if s.source[s.pos] == '[' {
		s.pos++
		for !s.eof() && s.source[s.pos] != ']' {
			s.pos++
		}
		if s.eof() {
			return "", newValidationError("unterminated bracketed identifier", start)
		}
		s.pos++
		return s.source[start:s.pos], nil
	}

	r := rune(s.source[s.pos])
	if !unicode.IsLetter(r) && s.source[s.pos] != '_' {
		return "", newValidationError("invalid identifier", start)
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
