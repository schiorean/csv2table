# csv2table

A flexible command line tool to automate parsing and importing of CSV files into database tables. At the moment only MySQL is supported. 

## Use case 

Every night you have to fetch a set of CSV files from a remote SFTP server and import each file into a local mysql table. If the destination table doesn't exist, automatically create it.

There may be several columns that need parsing special, for example column `expire_date` needs to be parsed like this:
1. normal values are formatted as `dd.mm.yyyy`
1. there is a special value that signifies that there's no upper limit in time, this value is sent as `31.12.2099`. This value you need to save it as a mysql `NULL` value.

## Documentation

### Configuration files

All configuration is manually defined in `toml` files. There are 2 types of configuration files
1. global configuration file named `csv2table.toml`
1. csv file specific configuration file, having the same name as the csv file but with ".toml" extension. The configuration option defined in this files overwrite the configuration options defined in the global file.

Each file is importing according to the configuration. First it is loaded the global configuration from `csv2table.toml` file, then the csv specific configuration is merged into the configuration. 

### Configuration options for MySQL

| Option | Description | Default value|
|---|---|---|
|host|mysql host name||
|port|mysql port|3306|
|etc|||


## TODO

1. Testing
1. Numeric column formatting

### Import parsing

Candidates
1. tengo https://github.com/d5/tengo --- most promising
1. lua https://github.com/yuin/gopher-lua
1. starlark https://github.com/google/starlark-go

## Notifications

1. Email notifications when import is done, or failed
1. slack?
