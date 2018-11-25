// Package csv2table provides a way to import csv files to corresponding database tables
// while providing different way to convert csv data to match your database definition.
//
// Currently it provides mysql implementation only.
package csv2table

// Config holds the global app configuration
// See newConfig() for default values
type Config struct {
	Verbose bool // whether to log various exection steps

	Table   string
	Mapping map[string]ColumnMapping
}

// ColumnMapping holds configuration of a csv column
type ColumnMapping struct {
	Type  string
	Index bool
}

/**
TODO:
	- generic config: the whole config, including csv2table.toml is read by the service
	- general csv2table.toml must be loaded only once
	- implement viper logic directly in service implementation and remove it from main
	- CHECK: is it possible to unmarshall to a not defined struct at compilte time? then we can unmarshall everything in main

*/

// DbService is the interface that needs to be implemented by various databases
type DbService interface {
	Start(fileName string, header []string) error
	End() error
	ProcessLine(line []string) error
	Test(configFile string) error
}

// NewConfig creates a new Config and applies defaults
func NewConfig() Config {
	c := Config{}

	return c
}
