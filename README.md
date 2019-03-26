# csv2table

A flexible command line tool to automate parsing of CSV files and importing them into database tables.

## Use case 

Every night you have to fetch a set of CSV files from a remote SFTP server and import each file into a local mysql table. If the destination table doesn't exist, automatically create it.

There may be several columns that need parsing, for example column `created_at` needs to be parsed like this:
1. normal values are formatted as `dd.mm.yyyy`
1. there is a special value that signifies that there's no upper limit in time, this value is sent as `31.12.2099`. This value you need to save it as a mysql `NULL` value.

## Documentation

All configuration is manually defined in `toml` files. There are 2 types 

## TODO
---

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
