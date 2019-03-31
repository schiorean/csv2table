package mysql

import (
	"regexp"
	"strings"
	"time"
)

// global list of prepared float parsers found in configuration files
var floatParsers map[string]*strings.Replacer

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
	case typeFloat:
		ret = parseFloat(format, value)
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

// parseFloat parses a float column from an unknown locale to system locale.
// The algorithtm is simple: the last non-numeric character in format string is considered the decimal point
func parseFloat(format string, value string) string {
	parser, exists := floatParsers[format]
	if !exists {
		// find kast non-numeric character => decimal point
		re := regexp.MustCompile("[^0-9]")
		match := re.FindAllString(format, -1)

		if len(match) == 0 {
			return value
		}

		// decimal point is the last element
		dp := match[len(match)-1]
		if dp == "," {
			parser = strings.NewReplacer(
				".", "",
				",", ".",
			)
		} else {
			parser = strings.NewReplacer(
				",", "",
			)
		}

		if floatParsers == nil {
			floatParsers = make(map[string]*strings.Replacer)
		}
		floatParsers[format] = parser
	}

	return parser.Replace(value)
}
