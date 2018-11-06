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
	db     *sqlx.DB             // mysql connection
	config csv2table.FileConfig // current active csv2table config
}

// Start initializes the processing of a csv2table.CsvFile
func (s *DbService) Start(file string, config csv2table.FileConfig) error {
	s.config = config

	err := s.connect()
	if err != nil {
		return err
	}

	exists, err := s.tableExists()
	if !exists {

	}

	fmt.Println(exists)

	return nil
}

// End finishes the processing of a csv2table.CsvFile
func (s *DbService) End() {
	if s.db != nil {
		s.db.Close()
		s.db = nil
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

// connect connects to the database
func (s *DbService) connect() error {
	var err error
	s.db, err = sqlx.Open("mysql", s.config.Username+":"+s.config.Password+"@/"+s.config.Db)
	if err != nil {
		return err
	}

	// ping it, to make sure db details are valid
	return s.db.Ping()
}

// tableExists check if a table exists
func (s *DbService) tableExists() (bool, error) {

	var err = s.db.QueryRowx("SHOW TABLES LIKE '" + s.config.Table + "'").Scan()
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	// if table doesn't exist => create it
	if err == sql.ErrNoRows {
		return false, nil
	}

	return true, nil
}
