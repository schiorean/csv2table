// Package mysql implements the mysql persistence interface of the csv2table package
package mysql

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"

	"github.com/schiorean/csv2table"
)

// DbService represents a service that implements csv2table.DbService for mysql
type DbService struct {
	DB *sqlx.DB
}

// Start initializes the processing of a csv2table.CsvFile
func (s *DbService) Start(file string, config csv2table.FileConfig) error {
	var err error
	s.DB, err = sqlx.Open("mysql", config.Username+":"+config.Password+"@/"+config.Db)
	if err != nil {
		return err
	}

	// ping it, to make sure db details are valid
	err = s.DB.Ping()
	if err != nil {
		return err
	}

	s.initStructure(config)

	return nil
}

func (s *DbService) initStructure(config csv2table.FileConfig) error {
	// var cnt int
	var err = s.DB.QueryRowx("SHOW TABLES LIKE '" + config.Table + "'").Scan()
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	fmt.Println(err)

	// if table doesn't exist => create it
	if err == sql.ErrNoRows {

	}

	// s.DB.Get()
	return nil
}

// End finishes the processing of a csv2table.CsvFile
func (s *DbService) End() {
	if s.DB != nil {
		s.DB.Close()
	}
}

// ProcessHeader processes the header of the csv file
func (s *DbService) ProcessHeader([]string) error {
	return nil
}

// ProcessLine processes a line header of the csv file
func (*DbService) ProcessLine([]string) error {
	return nil
}

// NewService creates a new instance of the DbService
func NewService() (*DbService, error) {
	return &DbService{}, nil
}
