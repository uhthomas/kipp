# conf
[![GoDoc](https://godoc.org/github.com/6f7262/conf?status.svg)](https://godoc.org/github.com/6f7262/conf)

## Installation and usage
```
go get github.com/6f7262/conf/cmd/conf
cp -r $GOPATH/src/github.com/6f7262/conf/default conf
cd conf
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
# make sure to run conf separately to openssl
# use "--proxy-header X-Real-IP" if being used behind a proxy for IP logging
conf --mime="mime.json"
```

## Help
### uploading via curl
```
curl -F "file=@<file>" https://conf.6f.io/upload
```
To upload to conf using its private feature, the CLI or web interface should be used.
```
$ conf --help
usage: conf [<flags>] <command> [<args> ...]

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).

Commands:
  help [<command>...]
	Show help.

  serve* [<flags>]
	Start a conf server.

  upload [<flags>] <file>
	Upload a file.
```
```
$ conf help serve
usage: conf serve [<flags>]

Start a conf server.

Flags:
  --help                       Show context-sensitive help (also try --help-long
                               and --help-man).
  --addr=":1337"               Server listen address.
  --insecure                   Disable https.
  --cert="cert.pem"            TLS certificate path.
  --key="key.pem"              TLS key path.
  --cleanup-interval=5m        Cleanup interval for deleting expired files.
  --mime=PATH                  A json formatted collection of extensions and
                               mime types.
  --driver="sqlite3"           Available database drivers: mysql, postgres,
                               sqlite3 and mssql.
  --driver-username="conf"     Database driver username.
  --driver-password=PASSWORD   Database driver password.
  --driver-path="conf.db"      Database driver path. ex: localhost:1337
  --expiration=24h             File expiration time.
  --max=150MB                  The maximum file size for uploads.
  --files="files"              File path.
  --tmp="files/tmp"            Temp path for in-progress uploads.
  --public="public"            Public path for web resources.
  --proxy-header=PROXY-HEADER  HTTP header to be used for IP logging if set.

```
```
$ conf help upload
usage: conf upload [<flags>] <file>

Upload a file

Flags:
  --help                    Show context-sensitive help (also try --help-long
                            and --help-man).
  --insecure                Don't verify SSL certificates.
  --private                 Encrypt the uploaded file.
  --url=https://conf.6f.io  Source URL.

Args:
  <file>  File to be uploaded
```

## Notes
* `Private` is not supported in IE, Edge or Safari due to incomplete implementation of the WebCrypto API. (this is however likely to change in the future when the frontend is rewritten in gopherjs)
* FilePath and TempPath must be located on the same drive as conf uploads files to its TempPath and then will move that file to the FilePath.
* It's recommended that extra mime types are used. This can be done by running conf with `--mime /path/to/mime.json`
* By default conf will use the remote address to log IP addresses. If conf is being run behind a proxy ensure the proxy is configured correctly and use `--proxy-header <header>`.
* For performance critical servers it's recommended to use nginx as a proxy to serve content. For instance, conf will only handle requests for content and uploading whereas nginx will handle serving static files such as its index, js or css. nginx configuration snippet:
```conf
server {
	server_name conf.6f.io;
	listen 80;

	root ~/conf/public;

	try_files $uri $uri/ @proxy;

	location @proxy {
		client_max_body_size    150m;
		proxy_redirect          off;
		proxy_set_header        Host            $host;
		proxy_set_header        X-Real-IP       $remote_addr;
		proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header        Upgrade         $http_upgrade;
		proxy_set_header        Connection      $http_connection;
		client_body_buffer_size 128k;
		proxy_connect_timeout   90;
		proxy_read_timeout      31540000;
		proxy_send_timeout      31540000;
		proxy_buffers           32 4k;
		proxy_buffering         off;
		proxy_request_buffering off;
		proxy_http_version      1.1;
		proxy_ssl_verify        off;
		proxy_pass              https://127.0.0.1:1337;
	}
}
```

## TODO
* Rewrite the frontend in gopherjs.
* Compress JS and CSS.
* Write tests.
* Cleanup frontend.
* Add context menus to files.
* Improve efficient of encryption / decryption. Web browsers are terrible and don't use proper Readers/Writers :(