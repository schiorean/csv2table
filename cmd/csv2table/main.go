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
	// iterate through all files in the directory
	// and find all csv files that have a matching configuration file (.toml)
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}

	found := false
	for _, f := range files {
		if strings.HasSuffix(strings.ToLower(f.Name()), ".csv") {
			found = true

			// mysql service, for now
			service := mysql.NewService()

			err = processCsv(service, f.Name())
			if err != nil {
				log.Printf("error while processing %s, %v", f.Name(), err)
			}
		}
	}

	if !found {
		log.Println("no files found")
	}
}

// processCsv reads a a csv file and imports it into a database table with similar structure
func processCsv(service csv2table.DbService, fileName string) error {
	v, err := getViper(fileName)
	if err != nil {
		return err
	}

	// initialize service
	err = service.Start(fileName, v)
	if err != nil {
		return err
	}
	defer service.End()

	// all good, now start csv file processing
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ';'

	// first line is always the header
	header, err := r.Read()
	if err != nil {
		return err
	}
	err = service.ProcessHeader(header)
	if err != nil {
		return err
	}

	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// rest of the lines are content
		err = service.ProcessLine(line)
		if err != nil {
			return err
		}
	}

	// signal end of csv file
	err = service.End()
	if err != nil {
		return err
	}

	return nil
}

// getViper initializes a new Viper instance merging the main config with the file based config
func getViper(fileName string) (*viper.Viper, error) {
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

	// next, try load file based config file
	baseName := strings.Replace(fileName, ".csv", "", -1)
	configFile := baseName + ".toml"

	if _, err := os.Stat(configFile); err == nil {
		if v == nil {
			v = viper.New()
		}
		v.SetConfigFile(configFile)

		err := v.MergeInConfig()
		if err != nil {
			return nil, fmt.Errorf("unable to read config file (%s), %v", configFile, err)
		}
	}

	// at least either main or file based configuration must be available
	if v == nil {
		return nil, fmt.Errorf("no configuration files found")
	}

	return v, nil
}
