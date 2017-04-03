# conf
A secure, temporary file storage site.

## Usage
```
$ go get github.com/6f7262/conf
$ mkdir conf
$ cd conf
$ conf
```

### Help
```
$ conf --help
usage: conf [<flags>]

Flags:
  --help                        Show context-sensitive help (also try --help-long and
                                --help-man).
  --addr=":1337"                Server listen address.
  --driver="sqlite3"            Available database drivers: mysql, postgres, sqlite3 and mssql
  --driver-source="conf.db"     Database driver source. mysql example: user:pass@/database
  --max=150MB                   The maximum file size
  --file-path="data/files"      The path to store uploaded files
  --temp-path="data/files/tmp"  The path to store uploading files
  --public-path="public"        The path where web resources are located.
```