// Package mysql implements the mysql persistence interface of the csv2table package
package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"

	"github.com/schiorean/csv2table"
)

// Config holds mysql specific configuration
type Config struct {
	Db       string
	Host     string // db host
	Port     int    // db port
	Username string // db username
	Password string // db password

	Table   string
	Mapping map[string]ColumnMapping

	Drop     bool // drop table if already exists?
	Truncate bool // truncate table before insert?

	AutoPk         bool   // use auto increment primary key?
	DefaultColType string // column type definintion
	TableOptions   string // default table options
	BulkInsertSize int    // how many rows to insert at once

	Verbose bool // whether to log various exection steps
}

// ColumnMapping holds configuration of a csv column
type ColumnMapping struct {
	Type  string
	Index bool
}

// config default options
const (
	defaultVerbose        = false
	defaultDrop           = false
	defaultTruncate       = false
	defaultBulkInsertSize = 5000

	autoPkType  = "`idauto` INT(11) NOT NULL AUTO_INCREMENT"
	autoPkIndex = "PRIMARY KEY(`idauto`)"
	colIndexTpl = "INDEX `{col}` (`{col}`)"
)

var sqlEscapeReplacer *strings.Replacer

// DbService represents a service that implements csv2table.DbService for mysql
type DbService struct {
	db *sqlx.DB // mysql connection

	fileName string // name of currently processed file
	config   Config // config for this file

	cols       []string // column names for current file
	rowCount   int      // number of rows currently processed
	statements []string // current list of sql statements, one for each row
}

// newConfig creates a new Config and applies defaults
func newConfig() Config {
	c := Config{
		Verbose:        defaultVerbose,
		Drop:           defaultDrop,
		Truncate:       defaultTruncate,
		BulkInsertSize: defaultBulkInsertSize,
	}

	return c
}

// NewService creates a new instance of the DbService
func NewService() *DbService {
	return &DbService{}
}

// Start initializes the processing of a csv file
func (s *DbService) Start(fileName string, v *viper.Viper) error {
	s.fileName = fileName

	// read config
	s.config = newConfig()

	// default table name is csv file name
	baseName := strings.Replace(fileName, ".csv", "", -1)
	s.config.Table = csv2table.SanitizeName(baseName)

	if v != nil {
		v.Unmarshal(&s.config)
	}

	if s.config.Verbose {
		log.Printf("Start importing %s\n", fileName)
	}

	err := s.connect()
	if err != nil {
		return err
	}

	// escape all names (columns and table name)
	s.config.Table = s.escapeString(s.config.Table)

	// allocate statements slice
	s.statements = make([]string, 0, s.config.BulkInsertSize)

	// initial row count
	s.rowCount = 0

	return nil
}

// End finishes the processing of a csv2table.CsvFile
func (s *DbService) End() error {
	defer func() {
		if s.db != nil {
			s.db.Close()
			s.db = nil
		}
	}()

	// insert any outstanding rows
	if len(s.statements) > 0 {
		err := s.insertOutstandingRows()
		if err != nil {
			return err
		}
	}

	return nil
}

// ProcessHeader is called to process the header, after Start() and before first call of ProcessLine()
func (s *DbService) ProcessHeader(header []string) error {
	// extract columns names from header
	s.cols = csv2table.SanitizeNames(header)
	s.cols = s.escapeStrings(s.cols)

	// escape all names (columns and table name)
	s.config.Table = s.escapeString(s.config.Table)
	s.cols = s.escapeStrings(s.cols)

	// prepare table
	exists, err := s.tableExists()
	if err != nil {
		return err
	}

	if exists && s.config.Drop {
		// DROP table
		if s.config.Verbose {
			log.Printf("Dropping table %v\n", s.config.Table)
		}

		_, err = s.db.Exec("drop table " + s.config.Table)
		if err != nil {
			return err
		}

		exists = false

	} else if s.config.Truncate {
		// TRUNCATE table
		if s.config.Verbose {
			log.Printf("Truncating table %v\n", s.config.Table)
		}

		_, err = s.db.Exec("truncate table " + s.config.Table)
		if err != nil {
			return err
		}
	}

	// create table if not exists
	if !exists {
		err = s.createTable()
		if err != nil {
			return err
		}
	}

	// allocate statements slice
	s.statements = make([]string, 0, s.config.BulkInsertSize)

	// initial row count
	s.rowCount = 0

	if s.config.Verbose {
		log.Printf("Starting import\n")
	}

	return nil
}

// ProcessLine processes a line header of the csv file
func (s *DbService) ProcessLine(line []string) error {
	data := make([]string, 0, len(s.cols))

	for _, v := range line {
		data = append(data, v)
		// mapping, exists := s.config.Mapping[colName]
		// if exists {
		// 	if len(mapping.Type) > 0 {
		// 		sqlType = mapping.Type
		// 	}
		// }
	}

	s.statements = append(s.statements, s.getSqlStringForRow(data))

	if len(s.statements) == s.config.BulkInsertSize {
		err := s.insertOutstandingRows()
		if err != nil {
			return err
		}
	}

	return nil
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
	var exists string
	var err = s.db.QueryRowx(fmt.Sprintf("SHOW TABLES LIKE '%v'", s.config.Table)).Scan(&exists)
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
	if s.config.Verbose {
		log.Printf("Creating table %v\n", s.config.Table)
	}

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

	_, err := s.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

// getColMapping creates the sql snippet for a column definition
// if not defined, use the default mapping
func (s *DbService) getColMapping(col string) ColumnMapping {
	mapping, exists := s.config.Mapping[col]
	if !exists {
		mapping = ColumnMapping{Type: s.config.DefaultColType, Index: false}
	}

	return mapping
}

// getSqlStringForRow creates an sql values string for insert
func (s *DbService) getSqlStringForRow(data []string) string {
	for i, v := range data {
		sqlv := "null"
		if v != "" {
			sqlv = fmt.Sprintf("'%s'", s.escapeString(v))
		}

		data[i] = sqlv
	}

	return "(" + strings.Join(data, ",") + ")\n"
}

// insertOutstandingRows inserts to db all collected rows up to this point
func (s *DbService) insertOutstandingRows() error {
	if len(s.statements) == 0 {
		return nil
	}

	if s.config.Verbose {
		log.Printf("Insert %v rows\n", len(s.statements))
	}

	cols := strings.Join(s.cols, ",")
	data := strings.Join(s.statements, ",")

	sql := fmt.Sprintf("insert into `%v` (%v) values\n %v", s.config.Table, cols, data)

	_, err := s.db.Exec(sql)
	if err != nil {
		return err
	}

	// empty statements
	s.statements = s.statements[:0]
	return nil
}
