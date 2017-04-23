### conf
A secure, temporary file storage site.

## Features
* Secure - files are encrypted and stored on the disk.
* Lightweight - Files aren't read into memory to be processed.
* Fast - files are uploaded, hashed and encrypted at the same time.
* Data deduplication - files are only stored once.
* Self-sufficient - files are automatically deleted after their expiration.

## Installation and usage
```
$ go get github.com/6f7262/conf/cmd/conf
$ cp -r $GOPATH/src/github.com/6f7262/conf/default conf
$ cd conf
$ conf --mime="mime.json"
```

## Help
```
$ conf --help
usage: conf [<flags>]

Flags:
  --help                     Show context-sensitive help (also try --help-long and --help-man).
  --addr=":1337"             Server listen address.
  --cleanup-interval=5m      Cleanup interval for deleting expired files.
  --driver="sqlite3"         Available database drivers: mysql, postgres, sqlite3 and mssql.
  --driver-source="conf.db"  Database driver source. mysql example: user:pass@/database.
  --mime="mime.json"         A json formatted collection of extensions and mime types.
  --max=150MB                The maximum file size limit for uploads.
  --file-path="files"        The path to store uploaded files.
  --temp-path="files/tmp"    The path to store uploading files.
  --public-path="public"     The path where web resources are located.

```

## Notes
* FilePath and TempPath must be located on the same drive as conf uploads files to it's TempPath and then will move that file to the FilePath.
* It is **highly** Recommended that extra mime types are added as go's standard set of types is very limited. This can be done by running conf with `--mime /path/to/mime.json`
* For performance critical servers it's recommended to use nginx as a proxy to serve content. For instance, conf will only handle requests for content and uploading whereas nginx will handle serving static files such as its index, js or css. nginx configuration snippet:
```
server {
    server_name conf.6f.io;
    listen 80;

    client_max_body_size 150m;

    root ~/conf/public;

    try_files $uri $uri/ @proxy;

    location @proxy {
        proxy_pass http://127.0.0.1:1337;
    }
}
```

## TODO
* Rewrite the frontend in gopherjs.
* Write tests.