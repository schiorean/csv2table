// Package csv2table provides a way to import csv files to corresponding mysql tables
// while providing different way to convert csv data to match your database definition
package csv2table

// global Config defaults
const (
	defaultHost           = "localhost"
	defaultPort           = 3306
	defaultDrop           = true
	defaultTruncate       = false
	defaultAutoPk         = false
	defaultAutoPkDef      = "id INT(11) NOT NULL AUTO_INCREMENT"
	defaultColDef         = "VARCHAR(100) NULL DEFAULT NULL"
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

	AutoPk    bool   // use auto increment primary key?
	AutoPkDef string // auto incremennt primary key definition

	DefaultColDef  string // default column definintion
	Drop           bool   // drop table if already exists?
	Truncate       bool   // truncate table before insert?
	BulkInsertSize int    // how many rows to insert at once
}

// FileConfig holds configuration associated with a csv file
// It embeds the global config and can overwrite it if needed
type FileConfig struct {
	Config
	Table   string
	Mapping map[string]FileMapping
}

// FileMapping holds configuration of a csv column
type FileMapping struct {
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
		Host:           defaultHost,
		Port:           defaultPort,
		AutoPk:         defaultAutoPk,
		AutoPkDef:      defaultAutoPkDef,
		DefaultColDef:  defaultColDef,
		Truncate:       defaultTruncate,
		BulkInsertSize: defaultBulkInsertSize,
	}

	return c
}
