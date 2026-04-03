package sqlsanitize

import "strings"

func RemoveComments(sql string) string {
	var builder strings.Builder

	for i := 0; i < len(sql); {
		if sql[i] == '\'' {
			start := i
			i++

			for i < len(sql) {
				if sql[i] == '\'' {
					if i+1 < len(sql) && sql[i+1] == '\'' {
						i += 2
						continue
					}

					i++
					break
				}

				i++
			}

			builder.WriteString(sql[start:i])
			continue
		}

		if i+1 < len(sql) && sql[i] == '-' && sql[i+1] == '-' {
			i += 2

			for i < len(sql) && sql[i] != '\n' && sql[i] != '\r' {
				i++
			}

			continue
		}

		if i+1 < len(sql) && sql[i] == '/' && sql[i+1] == '*' {
			i += 2

			for i+1 < len(sql) && !(sql[i] == '*' && sql[i+1] == '/') {
				i++
			}

			if i+1 < len(sql) {
				i += 2
			} else {
				i = len(sql)
			}

			continue
		}

		builder.WriteByte(sql[i])
		i++
	}

	return builder.String()
}
