// Package csv2table provides a way to import csv files to corresponding database tables
// while providing different way to convert csv data to match your database definition.
//
// Currently it provides mysql implementation only.
package csv2table

import (
	"fmt"

	"github.com/spf13/viper"
)

// DbService is the interface that needs to be implemented by various databases in order to offer csv2table support
type DbService interface {
	// Start is the first method called when a new csv file is processed.
	// The main configuration (csv2table.toml) is merged with file based configuration and passed as the viper.Viper param
	Start(fileName string, v *viper.Viper) error

	// End is called after the csv file has beed completely processed
	End() error

	// ProcessHeader is called for the 1st line of the csv file
	ProcessHeader(header []string) error

	// ProcessLine is called for each subsequent csv line
	ProcessLine(line []string) error
}

// ImportFileStatus holds each imported file status
type ImportFileStatus map[string]error

// UnmarshallConfig reads generic (non db provider) configuration
func UnmarshallConfig(v *viper.Viper) error {
	// email configuration
	err := v.UnmarshalKey("email", &emailConfig)
	if err != nil {
		return fmt.Errorf("unable to unmarshall loaded configuration email, %v", err)
	}

	return nil
}

// AfterImport is called after all files were processed
func AfterImport(status ImportFileStatus) {
	fmt.Println("status...")
}
