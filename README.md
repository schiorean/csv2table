# csv2table

A flexible command line tool to automate parsing and importing of CSV files into database tables. 

Currently the only database supported is MySQL.

## Use case 

Every night you have to fetch a set of CSV files from a remote SFTP server and import each file into a local mysql table. If the destination table doesn't exist, automatically create it.

There may be several columns that require special parsing:
* `expire_date`: normal values are sent as `dd.mm.yyyy`, but there is a special value that signifies that there's no upper limit in time, `31.12.2099`. This value must be saved as a database `NULL` value
* `amount`: convert from a specific locale (e.g. `12,5`) to the database locale (e.g. `12.5`)
* `start_date`: a more complex example, this column can have 2 types of values: 1) `dd.mm.yyyy` specifying an exact date or 2) an integer `n` meaning the number of months from January 1st of current year. In the later case we must calculate the the exact date and save it in database 

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

1. Auto PK custom name and definition
1. Email notifications
1. Tengo parsing

## Planned features 

1. slack notifications?
