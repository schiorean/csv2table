// Package csv2table provides a way to import csv files to corresponding mysql tables
// while providing different way to convert csv data to match your database definition
package csv2table

// global Config defaults (mysql compatible)
const (
	defaultDrop           = true
	defaultTruncate       = false
	defaultBulkInsertSize = 5000
)

// Config holds the global app configuration
// See newConfig() for default values
type Config struct {
	Db string

	Host     string // db host
	Port     int    // db port
	Username string // db username
	Password string // db password

	Drop     bool // drop table if already exists?
	Truncate bool // truncate table before insert?

	AutoPk         bool   // use auto increment primary key?
	DefaultColType string // column type definintion
	TableOptions   string // default table options
	BulkInsertSize int    // how many rows to insert at once

	Verbose bool // whether to log various exection steps
}

// FileConfig holds configuration associated with a csv file
// It embeds the global config and can overwrite it if needed
type FileConfig struct {
	Config
	Table   string
	Mapping map[string]ColumnMapping
}

// ColumnMapping holds configuration of a csv column
type ColumnMapping struct {
	Type  string
	Index bool
}

// DbService is the interface that needs to be implemented by various databases
type DbService interface {
	Start(config FileConfig, header []string) error
	End()
	ProcessLine(line []string) error
}

// NewConfig creates a new Config and applies defaults
func NewConfig() Config {
	c := Config{
		Drop:           defaultDrop,
		Truncate:       defaultTruncate,
		BulkInsertSize: defaultBulkInsertSize,
	}

	return c
}
