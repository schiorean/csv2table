package mysql

import "time"

// parseDate parses a date string using time.Parse(), 
// and returns it as a mysql valid date string
func parseDate(format string, in string) (string, error) {
	t, err := time.Parse(format, in)
	if (err != nil) {
		return "", err
	}

	return t.Format("2006-01-02"), nil
}

// ParseDateTime parses a datetime string using time.Parse(), 
// and returns it as a mysql valid datetime string
func parseDateTime(format string, in string) (string, error) {
	t, err := time.Parse(format, in)
	if (err != nil) {
		return "", err
	}

	return t.Format("2006-01-02 15:04:05"), nil
}

// formatValue function applies a formatting rule to a value,
// and returns the formatted value
// Possible formatting rules:
// "date:rule" : parse a date/time into a mysql date/time string, 
// 				 by parsing it according to the layout as defined by time.Parse()
func formatValue(value string, rule string) (string, error) {
	return value, nil
}