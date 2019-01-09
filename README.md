# kipp
[![GoDoc](https://godoc.org/github.com/6f7262/kipp?status.svg)](https://godoc.org/github.com/6f7262/kipp)

## Installation and usage
```
go get github.com/6f7262/kipp/cmd/kipp
cp -r $GOPATH/src/github.com/6f7262/kipp/default kipp
cd kipp
kipp --mime="mime.json"
```

## Help
### uploading via curl
```
curl -F "file=@<path>" https://kipp.6f.io
```
To upload to kipp using its private feature, the CLI or web interface should be used.
```
$ kipp help
usage: kipp [<flags>] <command> [<args> ...]

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).

Commands:
  help [<command>...]
    Show help.

  serve* [<flags>]
    Start a kipp server.

  upload [<flags>] <file>
    Upload a file.

```
```
$ kipp help upload
usage: kipp upload [<flags>] <file>

Upload a file.

Flags:
  --help                    Show context-sensitive help (also try --help-long
                            and --help-man).
  --insecure                Don't verify SSL certificates.
  --private                 Encrypt the uploaded file
  --url=https://kipp.6f.io  Source URL

Args:
  <file>  File to be uploaded
```

## Notes
* Does not support IE.
* --files and --tmp must be located on the same drive as kipp uploads files to --tmp and then will move it to --files.
* It's recommended that extra mime types are used. This can be done by running kipp with `--mime /path/to/mime.json`
* It's recommended to use nginx as a proxy to serve public files. For instance, kipp will only handle requests for uploading and viewing uploaded files then nginx will handle serving static files such as its index, js or css. nginx configuration snippet:
```kipp
server {
	server_name kipp.6f.io;
	listen 80;
	
	client_max_body_size 150m;
	expires max;

	root ~/kipp/public;

	try_files $uri $uri/ @proxy;

	location @proxy {
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
		# required to ensure cached files do not exceed their expiration date.
		expires                 off;
		proxy_pass              https://127.0.0.1:443;
	}
}
```