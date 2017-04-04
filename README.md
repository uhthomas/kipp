# conf
A secure, temporary file storage site.

## Usage
```
$ go get github.com/6f7262/conf
$ mkdir conf
$ cp $GOPATH/src/github.com/6f7262/conf/public conf/public
$ cp $GOPATH/src/github.com/6f7262/conf/mime.json conf/mime.json
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
  --driver="sqlite3"            Available database drivers: mysql, postgres, sqlite3 and mssql
  --driver-source="conf.db"     Database driver source. mysql example: user:pass@/database
  --max=150MB                   The maximum file size
  --file-path="data/files"      The path to store uploaded files
  --temp-path="data/files/tmp"  The path to store uploading files
  --public-path="public"        The path where web resources are located.
```

### Notes
* FilePath and TempPath must be located on the same drive as conf uploads files to it's TempPath and then will move that file to the FilePath.
* It is **Highly** Recommended that extra mime types are added as go's standard set of types is very limited. This can be done by running conf with `--mime /path/to/mime.json`
* conf's default structure looks like:
```
conf/
	files/
		tmp/
	public/
	mime.json
```