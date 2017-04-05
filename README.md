# conf
A secure, temporary file storage site.

## Usage
```
$ go get github.com/6f7262/conf
$ cp -r $GOPATH/src/github.com/6f7262/conf/default conf
$ conf --mime="conf/mime.json"
```

### Help
```
$ conf --help
usage: conf [<flags>]

Flags:
  --help                        Show context-sensitive help (also try --help-long and
                                --help-man).
  --addr=":1337"                Server listen address.
  --cleanup-interval=5m         Cleanup interval for deleting expired file
  --driver="sqlite3"            Available database drivers: mysql, postgres, sqlite3 and mssql
  --driver-source="conf/conf.db"  
                                Database driver source. mysql example: user:pass@/database
  --mime="conf/mime.json"       A json formatted collection of extensions and mime types.
  --max=150MB                   The maximum file size
  --file-path="conf/files"      The path to store uploaded files
  --temp-path="conf/files/tmp"  The path to store uploading files
  --public-path="conf/public"   The path where web resources are located.
```

### Notes
* FilePath and TempPath must be located on the same drive as conf uploads files to it's TempPath and then will move that file to the FilePath.
* It is **Highly** Recommended that extra mime types are added as go's standard set of types is very limited. This can be done by running conf with `--mime /path/to/mime.json`