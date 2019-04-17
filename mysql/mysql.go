// Package mysql implements the mysql persistence interface of the csv2table package
package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"

	"github.com/schiorean/csv2table"
)

// config default options
const (
	defaultPort = 3306

	defaultVerbose        = false
	defaultDrop           = false
	defaultTruncate       = false
	defaultBulkInsertSize = 5000
	defaultColType        = "VARCHAR(255) NULL DEFAULT NULL"
	defaultTableOptions   = "COLLATE='utf8_general_ci' ENGINE=InnoDB"

	autoPkColType  = "`id` INT(11) NOT NULL AUTO_INCREMENT"
	autoPkColIndex = "PRIMARY KEY(`id`)"
	colIndexTpl    = "INDEX `{col}` (`{col}`)"
)

// column types as understood by us
const (
	typeString   = "string"
	typeInt      = "int"
	typeFloat    = "float"
	typeDate     = "date"
	typeDateTime = "dateTime"
)

// Config holds mysql specific configuration
type Config struct {
	Db       string
	Host     string // db host
	Port     int    // db port
	Username string // db username
	Password string // db password

	Table      string                   // table name
	Mapping    map[string]ColumnMapping // columns mapping
	ColumnType map[string]string        // kind of columns type as understood by us

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
	Type        string
	Index       bool
	Format      string
	NullIf      []string
	NullIfEmpty bool
}

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
		Port:           defaultPort,
		Verbose:        defaultVerbose,
		Drop:           defaultDrop,
		Truncate:       defaultTruncate,
		BulkInsertSize: defaultBulkInsertSize,
		DefaultColType: defaultColType,
		TableOptions:   defaultTableOptions,
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
		err := v.Unmarshal(&s.config)
		if err != nil {
			return fmt.Errorf("unable to unmarshall loaded configuration, %v", err)
		}
	}

	if s.config.Verbose {
		log.Printf("Start importing %s\n", fileName)
	}

	err := s.connect()
	if err != nil {
		return err
	}

	// escape all names (columns and table name)
	s.config.Table = escapeString(s.config.Table)

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
	s.cols = escapeStrings(s.cols)

	// prepare table
	exists, err := s.tableExists()
	if err != nil {
		return err
	}

	// DROP table
	if exists && s.config.Drop {
		if s.config.Verbose {
			log.Printf("Dropping table %v\n", s.config.Table)
		}

		_, err = s.db.Exec("drop table " + s.config.Table)
		if err != nil {
			return err
		}

		exists = false

	}

	// TRUNCATE table
	if exists && s.config.Truncate {
		if s.config.Verbose {
			log.Printf("Truncating table %v\n", s.config.Table)
		}

		_, err = s.db.Exec("truncate table " + s.config.Table)
		if err != nil {
			return err
		}
	}

	// CREATE table if not exists
	if !exists {
		err = s.createTable()
		if err != nil {
			return err
		}
	}

	// parse db types => our types
	err = s.parseAndSetDbTypes()
	if err != nil {
		return err
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
	var err error

	// final column values slice
	// use pointer in order to use nil  to describe mysql NULL
	data := make([]*string, 0, len(s.cols))

	for i, value := range line {
		col := s.cols[i]
		mysqlValue, err := s.formatColumn(col, value)
		if err != nil {
			return err
		}

		// add value
		data = append(data, mysqlValue)
	}

	s.statements = append(s.statements, s.getSqlStringForRow(data))
	if len(s.statements) == s.config.BulkInsertSize {
		err = s.insertOutstandingRows()
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
		sql += fmt.Sprintf("%v ,\n", autoPkColType)
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
		sql += fmt.Sprintf("%v, \n", autoPkColIndex)
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

// getSqlStringForRow creates an sql values string for insert
func (s *DbService) getSqlStringForRow(data []*string) string {
	mysqlData := make([]string, len(data))

	for i, value := range data {
		sqlv := "NULL"
		if value != nil {
			sqlv = fmt.Sprintf("'%s'", escapeString(*value))
		}

		mysqlData[i] = sqlv
	}

	return "(" + strings.Join(mysqlData, ",") + ")\n"
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

// getColMapping creates the sql snippet for a column definition
// if not defined, use the default mapping
func (s *DbService) getColMapping(col string) ColumnMapping {
	mapping, exists := s.config.Mapping[col]
	if !exists {
		mapping = ColumnMapping{Type: s.config.DefaultColType, Index: false}
	}

	// set required default fields if not set
	if mapping.Type == "" {
		mapping.Type = s.config.DefaultColType
	}

	return mapping
}

// parseAndSetDbTypes parses current table metadata and update Config.ColumnType with equivalent types understood by us
func (s *DbService) parseAndSetDbTypes() error {
	s.config.ColumnType = make(map[string]string)

	rInt := regexp.MustCompile("(?i)int|unsigned|bit|tinyint|smallint|mediumint")
	rFloat := regexp.MustCompile("(?i)float|double")
	rDate := regexp.MustCompile("(?i)date")
	rDateTime := regexp.MustCompile("(?i)datetime|timestamp")

	for _, col := range s.cols {
		mapping := s.getColMapping(col)

		if rInt.MatchString(mapping.Type) {
			s.config.ColumnType[col] = typeInt
		} else if rFloat.MatchString(mapping.Type) {
			s.config.ColumnType[col] = typeFloat
		} else if rDateTime.MatchString(mapping.Type) {
			s.config.ColumnType[col] = typeDateTime
		} else if rDate.MatchString(mapping.Type) {
			s.config.ColumnType[col] = typeDate
		} else {
			s.config.ColumnType[col] = typeString // default
		}
	}

	return nil
}
