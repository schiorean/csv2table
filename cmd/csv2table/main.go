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

var (
	// c is the global config singleton
	c = csv2table.NewConfig()
)

// main is the entry routine
func main() {
	Run(".")
}

// Run function is the main routine that starts the csv import process
func Run(directory string) {
	// load global config
	if _, err := os.Stat("csv2table.toml"); err == nil {
		v := viper.New()
		v.SetConfigFile("csv2table.toml")
		err := v.ReadInConfig()
		if err != nil {
			log.Fatalf("unable to read global config file, %v", err)
		}

		err = v.Unmarshal(&c)
		if err != nil {
			log.Fatalf("unable to decode global config into struct, %v", err)
		}
	}

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
		fmt.Println("no files found")
	}
}

// processCsv reads a a csv file and imports it into a database table with similar structure
func processCsv(service csv2table.DbService, fileName string) error {
	fileConfig, err := getFileConfig(fileName)
	if err != nil {
		return err
	}

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

	// initialize service
	err = service.Start(fileConfig, header)
	if err != nil {
		return err
	}
	defer service.End()

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

	return nil
}

// getFileConfig prepares a FileConfig associated with a csv file.
// Defaults to global config, and it may be overridden/extended by a csv file based config
func getFileConfig(fileName string) (csv2table.FileConfig, error) {
	fileConfig := csv2table.FileConfig{}

	// apply global config
	fileConfig.Config = c

	// default table name is the base name
	baseName := strings.Replace(fileName, ".csv", "", -1)
	fileConfig.Table = csv2table.SanitizeName(baseName)

	// check load a matching config file
	configFile := baseName + ".toml"
	if _, err := os.Stat(configFile); err == nil {
		v := viper.New()
		v.SetConfigFile(configFile)

		err := v.ReadInConfig()
		if err != nil {
			return csv2table.FileConfig{}, fmt.Errorf("unable to read config file (%s), %v", configFile, err)
		}

		// unmarshal file based global config. Why the "Config" embeded struct can't be unmarshaled automatically? No clue, so for now do it separately
		err = v.Unmarshal(&fileConfig.Config)
		if err != nil {
			return csv2table.FileConfig{}, fmt.Errorf("unable to decode config file  (%s), %v", configFile, err)
		}

		// unmarshal main config (table, mapping)
		err = v.Unmarshal(&fileConfig)
		if err != nil {
			return csv2table.FileConfig{}, fmt.Errorf("unable to decode config file (%s), %v", configFile, err)
		}
	}

	return fileConfig, nil
}
