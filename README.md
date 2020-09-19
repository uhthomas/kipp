# kipp
[![GoDoc](https://godoc.org/github.com/uhthomas/kipp?status.svg)](https://godoc.org/github.com/uhthomas/kipp)
[![Go Report Card](https://goreportcard.com/badge/github.com/uhthomas/kipp)](https://goreportcard.com/report/github.com/uhthomas/kipp)

## Getting started
The easiest way to get started with kipp is by using the image published to
[Docker Hub](https://hub.docker.com/repository/docker/uhthomas/kipp). The
service is then available simply by running:
```
docker pull uhthomas/kipp
docker run uhthomas/kipp
```

## Databases
Databases can be configured using the `--database` flag. The flag requires
the input be parsable as a URL. See the [url.Parse](https://golang.org/pkg/net/url/#Parse)
docs for more info.

### [Badger](https://github.com/dgraph-io/badger)
Badger is a fast, embedded database which is great for single instances.

### SQL
Kipp uses a generic SQL driver, but currently only loads:
* [PostgreSQL](https://www.postgresql.org/)

As long as a database supports Go's [sql](https://golang.org/pkg/database/sql/)
package, it can be used. Please file an issue for requests.

## File systems
File systems can be configured using the `--filesystem` flag. The flag requires
the input be parsable as a URL. See the [url.Parse](https://golang.org/pkg/net/url/#Parse)
docs for more info.

### Local (your local file system)
The local filesystem does not require any special formatting, and can be used
like a regular path such

```
--filesystem /path/to/files
```

### [AWS S3](https://aws.amazon.com/s3/)
AWS S3 requires the `s3` scheme, and has the following syntax:

```
--filesystem s3://some-token:some-secet@some-region/some-bucket?endpoint=some-endpoint.
```

The `region` and `bucket` are required.

The [user info](https://tools.ietf.org/html/rfc2396#section-3.2.2) section is
optional, if present, will create new static credentials. Otherwise, the default
AWS SDK credentials will be used.

The `endpoint` is optional, and will use the default AWS endpoint if not present.
This is useful for using S3-compatible services such as:
* [Google Cloud Storage](https://cloud.google.com/storage) - storage.googleapis.com
* [Linode Object Storage](https://www.linode.com/products/object-storage/) - linodeobjects.com
* [Backblaze B2](https://www.backblaze.com/b2/cloud-storage.html) - backblazeb2.com
* [DigitalOcean Spaces](https://www.digitalocean.com/products/spaces/) - digitaloceanspaces.com
* ... etc

#### Policy
Required actions:
* `s3:DeleteObject`
* `s3:GetObject`
* `s3:PutObject`

This is subject to change in future as more features are added.

## Building from source
Kipp builds, tests and compiles using [Bazel](https://bazel.build). To run/build
locally with bazel:
```
git clone git@github.com:uhthomas/kipp
cd kipp
bazel run //cmd/kipp
```

## API
Kipp has two main components; uploading files and downloading files. Files can
be uploaded by POSTing a multipart form to the `/` endpoint like so:
```
curl https://kipp.6f.io -F file="some content"
```
The service will then respond with a `302 (See Other)` status and the location
of the file. It will also write the location to the response body.

Kipp also serves all files located in the `web` directory by default, but can
either be disabled or changed to a different location.
