package mysql

import "time"

// parseDate parses a date string using time.Parse(),
// and returns it as a mysql valid date string
func parseDate(format string, value string) (string, error) {
	t, err := time.Parse(format, value)
	if err != nil {
		return "", err
	}

	return t.Format("2006-01-02"), nil
}

// parseDateTime parses a datetime string using time.Parse(),
// and returns it as a mysql valid datetime string
func parseDateTime(format string, value string) (string, error) {
	t, err := time.Parse(format, value)
	if err != nil {
		return "", err
	}

	return t.Format("2006-01-02 15:04:05"), nil
}

// formatColumn formats a column value based on the column type and provided format
func formatColumn(columnType string, format string, value string) (string, error) {
	var ret string
	var err error

	switch columnType {
	case typeDate:
		ret, err = parseDate(format, value)
	}

	return ret, err
}

// applyNull sets a column value to mysql NULL if the raw value matches a value from the nullIf slice
// return true if column should be NULL, false otherwise
func applyNull(nullIf []string, value string) bool {
	for _, nullMatch := range nullIf {
		if value == nullMatch {
			return true
		}
	}

	return false
}
