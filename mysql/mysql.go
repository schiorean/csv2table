// Package mysql implements the mysql persistence interface of the csv2table package
package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"

	"github.com/schiorean/csv2table"
)

const (
	autoPkType  = "`idauto` INT(11) NOT NULL AUTO_INCREMENT"
	autoPkIndex = "PRIMARY KEY(`idauto`)"
	colIndexTpl = "INDEX `{col}` (`{col}`)"
)

// DbService represents a service that implements csv2table.DbService for mysql
type DbService struct {
	db     *sqlx.DB             // mysql connection
	config csv2table.FileConfig // current active csv2table config
	cols   []string             // column names for current file
}

// Start initializes the processing of a csv2table.CsvFile
func (s *DbService) Start(config csv2table.FileConfig, header []string) error {
	s.config = config

	err := s.connect()
	if err != nil {
		return err
	}

	// extract columns names from header
	s.cols = csv2table.SanitizeNames(header)

	// prepare table
	exists, err := s.tableExists()
	if exists && s.config.Drop {
		_, err = s.db.Exec("drop table " + s.config.Table)
		if err != nil {
			return err
		}
		exists = false
	}

	// create table if not exists
	if !exists {
		err = s.createTable()
		if err != nil {
			return err
		}
	}

	return nil
}

// End finishes the processing of a csv2table.CsvFile
func (s *DbService) End() {
	if s.db != nil {
		s.db.Close()
		s.db = nil
	}
}

// ProcessLine processes a line header of the csv file
func (*DbService) ProcessLine([]string) error {
	return nil
}

// NewService creates a new instance of the DbService
func NewService() *DbService {
	return &DbService{}
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

	var err = s.db.QueryRowx(fmt.Sprintf("SHOW TABLES LIKE '%v'", s.config.Table)).Scan()
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	// if table doesn't exist => create it
	if err == sql.ErrNoRows {
		return false, nil
	}

	return true, nil
}

// createTable creates the destination table
func (s *DbService) createTable() error {
	sql := fmt.Sprintf("create table `%v` (\n", s.config.Table)

	// add auto-increment PK
	if s.config.AutoPk {
		sql += fmt.Sprintf("%v ,\n", autoPkType)
	}

	var indexes []string

	// add column definitions
	for _, col := range s.cols {
		mapping := s.getColMapping(col)
		sql += fmt.Sprintf("`%v` %v, \n", col, mapping.Type)

		// build indexes
		if mapping.Index {
			indexes = append(indexes, strings.Replace(colIndexTpl, "{col}", col, -1))
		}
	}

	// now add indexes
	if s.config.AutoPk {
		sql += fmt.Sprintf("%v, \n", autoPkIndex)
	}

	for _, index := range indexes {
		sql += fmt.Sprintf("%v, \n", index)
	}

	// remove last , and close columns definition
	sql = strings.TrimSuffix(sql, ", \n") + "\n)\n"

	// add table options
	sql += s.config.TableOptions

	fmt.Println(sql)
	return nil
}

// getColMapping creates the sql snippet for a column definition
func (s *DbService) getColMapping(col string) csv2table.ColumnMapping {
	mapping, exists := s.config.Mapping[col]
	if !exists {
		mapping = csv2table.ColumnMapping{Type: s.config.DefaultColType, Index: false}
	}

	return mapping
}
