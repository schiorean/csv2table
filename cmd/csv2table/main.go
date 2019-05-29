package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/schiorean/csv2table"
	"github.com/schiorean/csv2table/mysql"

	"github.com/spf13/viper"
)

// main is the entry routine
func main() {
	Run(".")
}

// Run function is the main routine that starts the csv import process
func Run(directory string) {
	var err error

	// iterate through all files in the directory
	// and find all csv files that have a matching configuration file (.toml)
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}

	// import status list collected from each processed file
	statuses := make([]csv2table.ImportFileStatus, 0, len(files))

	found := false
	for _, f := range files {
		if strings.HasSuffix(strings.ToLower(f.Name()), ".csv") {
			found = true

			// in order for a file to be processed it must have a matching configuration file
			if !hasConfigFile(f.Name()) {
				continue
			}

			// mysql service, for now
			service := mysql.NewService()

			rowCount, err := processCsv(service, f.Name())
			if err != nil {
				log.Fatalf("error while processing %s, %v", f.Name(), err)
			}

			// add import status
			statuses = append(statuses, csv2table.ImportFileStatus{
				FileName: f.Name(),
				Error:    err,
				RowCount: rowCount,
			})
		}
	}

	if !found {
		log.Println("no files found")
	}

	err = csv2table.AfterImport(statuses)
	if err != nil {
		log.Fatalf("error while running after import routine, %v", err)
	}
}

// processCsv reads a a csv file and imports it into a database table with similar structure
func processCsv(service csv2table.DbService, fileName string) (int, error) {
	v, err := getFileViper(fileName)
	if err != nil {
		return 0, err
	}

	// unmarshall generic (non db provider)  configuration
	if v != nil {
		err := csv2table.UnmarshallConfig(v)
		if err != nil {
			return 0, err
		}
	}

	// initialize service
	err = service.Start(fileName, v)
	if err != nil {
		return 0, err
	}
	defer service.End()

	// all good, now start csv file processing
	f, err := os.Open(fileName)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ';'

	// first line is always the header
	header, err := r.Read()
	if err != nil {
		return 0, err
	}
	err = service.ProcessHeader(header)
	if err != nil {
		return 0, err
	}

	rowCount := 0
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}

		rowCount++

		// rest of the lines are content
		err = service.ProcessLine(line)
		if err != nil {
			return rowCount, err
		}
	}

	// signal end of csv file
	err = service.End()
	if err != nil {
		return rowCount, err
	}

	return rowCount, nil
}

// getGlobalViper reads global viper configuration from csv2table.toml
func getGlobalViper() (*viper.Viper, error) {
	var v *viper.Viper

	// try load global config
	if _, err := os.Stat("csv2table.toml"); err == nil {
		v = viper.New()
		v.SetConfigFile("csv2table.toml")

		err := v.ReadInConfig()
		if err != nil {
			return nil, fmt.Errorf("unable to read global config file, %v", err)
		}
	}

	return v, nil
}

// getFileViper initializes a new Viper instance merging the main config with the file based config
func getFileViper(fileName string) (*viper.Viper, error) {
	var v *viper.Viper
	var err error

	// try load global config
	v, err = getGlobalViper()
	if err != nil {
		return nil, err
	}

	if !hasConfigFile(fileName) {
		return nil, nil
	}

	if v == nil {
		v = viper.New()
	}

	configFile := getConfigFileName(fileName)
	v.SetConfigFile(configFile)

	err = v.MergeInConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to read config file (%s), %v", configFile, err)
	}

	return v, nil
}

// getConfigFileName returns the matching config file name of a csv
func getConfigFileName(fileName string) string {
	return strings.Replace(fileName, ".csv", "", -1) + ".toml"
}

// hasConfigFile checks wether a csv file has an matching configuration file
func hasConfigFile(fileName string) bool {
	_, err := os.Stat(getConfigFileName(fileName))
	return err == nil
}
