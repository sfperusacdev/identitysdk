package sqlviews

import (
	"regexp"
	"strings"
)

func FindViewNames(sql string) []string {
	normalized := strings.ReplaceAll(sql, "\r", " ")
	normalized = strings.ReplaceAll(normalized, "\n", " ")
	normalized = strings.ReplaceAll(normalized, "\t", " ")

	reSpace := regexp.MustCompile(`\s+`)
	normalized = reSpace.ReplaceAllString(normalized, " ")

	ident := `(?:"[^"]+"|\w+)`

	re := regexp.MustCompile(`(?i)\bCREATE\s+(?:OR\s+REPLACE\s+)?VIEW\s+(` +
		ident +
		`(?:\s*\.\s*` +
		ident +
		`)?)`)

	matches := re.FindAllStringSubmatch(normalized, -1)

	out := make([]string, 0, len(matches))
	for _, m := range matches {
		name := strings.TrimSpace(strings.ReplaceAll(m[1], " ", ""))
		if name == "" {
			continue
		}
		out = append(out, name)
	}

	return out
}
