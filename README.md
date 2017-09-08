# conf
[![GoDoc](https://godoc.org/github.com/6f7262/conf?status.svg)](https://godoc.org/github.com/6f7262/conf)

A temporary file storage server.

## Installation and usage
```
go get github.com/6f7262/conf/cmd/conf
cp -r $GOPATH/src/github.com/6f7262/conf/default conf
cd conf
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
# make sure to run conf separately to openssl
conf --mime="mime.json" --secure
```

## Help
```
$ conf --help
usage: conf [<flags>]

Flags:
  --help                      Show context-sensitive help (also try --help-long and
                              --help-man).
  --addr=":1337"              Server listen address.
  --secure                    Enable https.
  --cert="cert.pem"           TLS certificate path.
  --key="key.pem"             TLS key path.
  --cleanup-interval=5m       Cleanup interval for deleting expired files.
  --mime=PATH                 A json formatted collection of extensions and mime types.
  --driver="sqlite3"          Available database drivers: mysql, postgres, sqlite3 and mssql.
  --driver-username="conf"    Database driver username.
  --driver-password=PASSWORD  Database driver password.
  --driver-path="conf.db"     Database driver path. ex: localhost:1337
  --expiration=24h            File expiration time.
  --max=150MB                 The maximum file size for uploads.
  --files="files"             File path.
  --tmp="files/tmp"           Temp path for in-progress uploads.
  --public="public"           Public path for web resources.
```

## Notes
* FilePath and TempPath must be located on the same drive as conf uploads files to its TempPath and then will move that file to the FilePath.
* It's recommended that extra mime types are used. This can be done by running conf with `--mime /path/to/mime.json`
* For performance critical servers it's recommended to use nginx as a proxy to serve content. For instance, conf will only handle requests for content and uploading whereas nginx will handle serving static files such as its index, js or css. nginx configuration snippet:
```
server {
    server_name conf.6f.io;
    listen 80;

    root ~/conf/public;

    try_files $uri $uri/ @proxy;

    location @proxy {
        client_max_body_size 150m;
        proxy_buffering off;
        proxy_request_buffering off;
        proxy_ssl_verify off;
        proxy_pass https://127.0.0.1:1337;
    }
}
```

## TODO
* Rewrite the frontend in gopherjs.
* Compress JS and CSS.
* Write tests.