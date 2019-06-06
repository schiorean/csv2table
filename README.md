# csv2table

A fast and flexible command line tool to automate parsing and importing of CSV files into database tables. Currently the only database supported is MySQL.

## Use case 

Every night you have to fetch a set of CSV files from a remote SFTP server and import each file into a corresponding mysql table. If the destination table doesn't exist, automatically create it.

There may be several columns that require special processing:
* `expire_date`: normal values are sent as `dd.mm.yyyy`, but there is a special value that signifies that there's no upper limit in time, `31.12.2099`. This value must be saved as a database `NULL` value
* `expire_date`: besides processing the value, an index is needed on this column for faster searching
* `amount`: convert from a specific locale (e.g. `12,5`) to the database locale (e.g. `12.5`)
* `start_date`: a more complex example, this column can have 2 types of values: 1) `dd.mm.yyyy` specifying an exact date or 2) an integer `n` meaning the number of months from January 1st of current year. In the later case we must calculate the the exact date and save it in database. (NOTE: Such complex example will be possible only after integrating an embedded language. See "Planned features" section).

## Documentation

### Table and column names transformations

Table and column names are sanitized and transformed before creating the corresponding database table. The following rules apply:
* white space is trimmed
* inner white space and `-` are replaced with `_`
* special characters that can be part of the CSV header but can't be part of a database name are removed: `<>/\()`
* "umlaut" characters are replaced by their normal counterparts (e.g. `ä` or `á` is replaced by `a`)
* and finally all names are lower cased.

Example: A CSV column named `Date of Receipt` will be saved in database as `date_of_receipt`.

### Configuration files

All configuration is manually defined in [`toml`](https://github.com/toml-lang/toml) files. There are 2 types of configuration files:

#### 1) Global configuration file `csv2table.toml`

If there is a file named `csv2table.toml` in the working directory (the directory where csv files are found) it is loaded as the first configuration file. You can skip using the global configuration file, its common usage is when you have many CSV files that share identical configuration options (e.g. the database credentials).

#### 2) CSV specifific configuration file

They have the same name as the csv file but with ".toml" extension. The main purpose of the CSV configuration file is to define the column mapping. Besides column mapping you can also define global configuration options in which case they will overwrite the configuration options defined in the global file. 

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
|`drop`|drop and recreate table before import (true/false)|false|
|`truncate`|truncate table before import (true/false)|false|
|`autoPk`|create an auto increment PK (int)|false|
|`defaultColType`|default column definition|`VARCHAR(255) NULL DEFAULT NULL`|
|`tableOptions`|table options when creating the table|`COLLATE='utf8_general_ci' ENGINE=InnoDB`|
|`bulkInsertSize`|how many rows to insert at once|10000|
|`verbose`|verbosity to console|false|
|`email`|a section where email notifications cand be configured, see "Email notifications" section||


### Column mapping

Column mapping is defined in the CSV specific configuration file. Mapping options for each column are grouped under a `mapping.column_name` table. If we take the example from the "Use case" section, mapping options for column `expire_date` will be grouped under `mapping.expire_date` table.

Note: the name of the column is the database column name (see "Table and column names transformations" section).

Column mapping options:

| Mapping option | Description | Example | Default value|
|---|---|---|---|
|`type`|column type|`type = "INT NULL DEFAULT NULL"`|defaults to global option `defaultColType`|
|`index`|add column index (true/false)|`index = true`|false|
|`nullIf`|set column to DB null if its value is one in the list|`nullIf = ["31.12.2999", ""]`||
|`nullIfEmpty`|set column to DB null if its value is empty string|`nullIfEmpty = true`|false|

`format` mapping option is used to format a value from the CSV format to DB format. Possible patterns:

| Usage | Description | Example|
|---|---|---|
|decimal point|hint decimal point by simply assigning a number containing the CSV decimal point|`format = "1,2"` (hint that "," is the decimal point)|
|date parsing|parse a date using "Go" language [date and time pattern matching](https://yourbasic.org/golang/format-parse-string-time-date-example/#basic-time-format-example) |`format = "02.01.2006"` (date format is dd.mm.yyyy)|
|time parsing|parse a date/time using "Go" language [date and time pattern matching](https://yourbasic.org/golang/format-parse-string-time-date-example/#basic-time-format-example) |`format = "02.01.2006 15:04:05"` (date format is dd.mm.yyyy hh:mm:ss)|

Example of mapping options for column `expire_date`:
```toml
[mapping.expire_date]
    type = "DATE NULL DEFAULT NULL"
    format = "02.01.2006"
    index = true
```

### Email notifications

It's possible to enable email notifications through SMTP protocol. Example sending notifications when an error occurs, usig GMail SMTP.
```toml
[email]
    sendOnSuccess = false # don't send email if everything ok
    sendOnError = true # send if an error occurs
    from = "me@gmail.com"
    to = ["me@gmail.com"]
    smtpServer = "smtp.gmail.com:587"
[email.plainAuth]
    identity = ""
    username = "me@gmail.com"
    password = "gmail_secret"
    host = "smtp.gmail.com"
```

TODO documentation: It's possible to overwrite the default emails subject and body as as configuration options.

### Full example

`sample_import.csv` file to be imported
```csv
No ID;Reading;Reading_Date;Channel
1;2,5;02.05.2014;X
2;2,5;02.05.2014;X
3;2,5;02.05.2014;First
4;2,5;02.05.2014;X
5;2,5;02.05.2014;Last
6;2,5;07.05.2014;X
7;2,5;17.07.2014;X
8;2,5;23.07.2014;X
9;2,5;28.07.2014;X
10;2,5;17.03.2015;X
11;2,5;01.02.2016;
```

`csv2table.toml` global configuration
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
    sendOnError = true # send email notification on error
    from = "myself@gmail.com"
    to = ["myself@gmail.com"]
    smtpServer = "smtp.gmail.com:587"
[email.plainAuth]
    identity = ""
    username = "myself@gmail.com"
    password = "gmail_secret"
    host = "smtp.gmail.com"
```

`sample_import.toml` file configuration
```toml
[mapping.no_id]
    type = "INT NULL DEFAULT NULL"
    index = true
[mapping.reading]
    type = "DOUBLE NULL DEFAULT NULL"
    format = "1,2" 
[mapping.reading_date]
    type = "DATE NULL DEFAULT NULL"
    format = "02.01.2006"
    nullIf = ["31.12.2999"]
[mapping.channel]
    nullIF = ["Last"]
    nullIfEmpty = true
```

And result in database

![result](https://raw.githubusercontent.com/schiorean/csv2table/master/doc/sample_import_result.png)

## Planned features 
1. Integrate an embedded language to allow complex culomn parsing. Candidates: Tengo, Lua.
