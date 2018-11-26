// Package mysql implements the mysql persistence interface of the csv2table package
// This file holds mysql escape routines
package mysql

import "strings"

var sqlEscapeReplacer *strings.Replacer

// escapeString escapes a string by adding backslashes before special
// characters, and turning others into specific escape sequences, such as
// turning newlines into \n and null bytes into \0.
// source https://github.com/go-sql-driver/mysql/blob/master/utils.go, escapeStringBackslash function
func (s *DbService) escapeString(us string) string {
	if sqlEscapeReplacer == nil {
		sqlEscapeReplacer = strings.NewReplacer(
			"\x00", "\\0",
			"\n", "\\n",
			"\r", "\\r",
			"\x1a", "\\Z",
			"'", "\\'",
			"\"", "\\\"",
			"\\", "\\\\",
		)
	}

	return sqlEscapeReplacer.Replace(us)
}

// escapeStrings applies escapeString over a list of strings
func (s *DbService) escapeStrings(us []string) []string {
	es := make([]string, len(us))
	for i, v := range us {
		es[i] = s.escapeString(v)
	}

	return es
}
