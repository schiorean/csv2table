package mysql

import "time"

// formatColumn formats a column value based on various mapping flags
func (s *DbService) formatColumn(col string, value string) (*string, error) {
	var err error

	mysqlValue := new(string)
	*mysqlValue = value

	mapping, exists := s.config.Mapping[col]
	if !exists {
		return mysqlValue, nil
	}

	// nullIfEmpty: set empty column as mysql NULL
	if mapping.NullIfEmpty && value == "" {
		mysqlValue = nil
	}

	// nullIf: set column as mysql NULL
	if mysqlValue != nil && len(mapping.NullIf) > 0 {
		if applyNull(mapping.NullIf, value) {
			mysqlValue = nil
		}
	}

	// format: apply value formatting
	if mysqlValue != nil && mapping.Format != "" {
		*mysqlValue, err = parseColumn(s.config.ColumnType[col], mapping.Format, *mysqlValue)
		if err != nil {
			return nil, err
		}
	}

	return mysqlValue, nil
}

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

// parseColumn parses a column value based on the column type and provided format
func parseColumn(columnType string, format string, value string) (string, error) {
	var ret string
	var err error

	switch columnType {
	case typeDate:
		ret, err = parseDate(format, value)
	case typeDateTime:
		ret, err = parseDateTime(format, value)
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
