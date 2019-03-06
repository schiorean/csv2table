# csv2table
A flexible command line tool to import a CSV file into a database table.

Currently the only supported database is MySql.

## TODO

### Column custom format

1. Dates
1. Numbers

### Import parsing

Candidates
1. tengo https://github.com/d5/tengo --- most promising
1. lua https://github.com/yuin/gopher-lua
1. starlark https://github.com/google/starlark-go

### Other
1. Add tests
1. Performance comparision with PHP script
1. Performance improvements (goroutines for reading from csv and writing to db?)

## Notifications

1. Email notifications when import is done, or failed
1. slack?
