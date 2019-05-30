# csv2table

A flexible command line tool to automate parsing and importing of CSV files into database tables. Currently the only database supported is MySQL.

## Use case 

Every night you have to fetch a set of CSV files from a remote SFTP server and import each file into a local mysql table. If the destination table doesn't exist, automatically create it.

There may be several columns that require special processing:
* `expire_date`: normal values are sent as `dd.mm.yyyy`, but there is a special value that signifies that there's no upper limit in time, `31.12.2099`. This value must be saved as a database `NULL` value
* `expire_date`: besides processing the value, an index is needed on this column for faster searching
* `amount`: convert from a specific locale (e.g. `12,5`) to the database locale (e.g. `12.5`)
* `start_date`: a more complex example, this column can have 2 types of values: 1) `dd.mm.yyyy` specifying an exact date or 2) an integer `n` meaning the number of months from January 1st of current year. In the later case we must calculate the the exact date and save it in database. (NOTE: Such complex example will be possible only after integrating an embedded language. See "Planned features" section).

## Usage


## Documentation

### Configuration files

All configuration is manually defined in [`toml`](https://github.com/toml-lang/toml) files. There are 2 types of configuration files:

#### 1) Global configuration file `csv2table.toml`

If there is a file named `csv2table.toml` in the working directory (the directory where csv files are found) it is loaded as the first configuration file. You can skip using the global configuration file, its usage is when you have many CSV files that have some identical configuration options (e.g. the database credentials).

#### 2) CSV file specifific configuration file

They have the same name as the csv file but with ".toml" extension. The configuration option defined in this files overwrite the configuration options defined in the global file.

**ATTENTION**: The file specific configuration file is mandatory, otherwise the CSV file will not be imported.

### Configuration options 

Main configuration options:

| Option | Description | Default value|
|---|---|---|
|`host`|mysql host name||
|`port`|mysql port|3306|
|`db`|database name||
|`username`|database username||
|`password`|database password||
|`table`|table name|defaults to "sanitized" name of the CSV file|
|`drop`|drop and recreate table before import (true|false)|false|
|`truncate`|truncate table before import (true|false)|false|
|`autoPk`|create an auto increment PK (int)|false|
|`defaultColType`|default column definition|`VARCHAR(255) NULL DEFAULT NULL`|
|`tableOptions`|table options when creating the table|`COLLATE='utf8_general_ci' ENGINE=InnoDB`|
|`bulkInsertSize`|how many rows to insert at once|10000|
|`verbose`|verbosity to console|false|
|`email`|a section where email||

### Column mapping



Column mapping options:


```toml
host = "localhost"
port = 3306
username = "my_username"
password = "my_pwd"
db = "my_db"

# config defaults
verbose = true
drop = true
truncate = true
autoPk = true
bulkInsertSize = 15000

# email 
[email]
    sendOnSuccess = false
    sendOnError = true
    from = "myself@gmail.com"
    to = ["myself@gmail.com"]
    smtpServer = "smtp.gmail.com:587"
[email.plainAuth]
    identity = ""
    username = "myself@gmail.com"
    password = "gmail_secret"
    host = "smtp.gmail.com"
```

## Planned features 
1. Integrate an embedded language to allow complex culomn parsing. Candidates: Tengo, Lua.
