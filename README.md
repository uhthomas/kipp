# kipp
[![GoDoc](https://godoc.org/github.com/uhthomas/kipp?status.svg)](https://godoc.org/github.com/uhthomas/kipp)
[![Go Report Card](https://goreportcard.com/badge/github.com/uhthomas/kipp)](https://goreportcard.com/report/github.com/uhthomas/kipp)

## Getting started
The easiest way to get started with kipp, is by using the image published to
[Docker Hub](https://hub.docker.com/repository/docker/uhthomas/kipp). The
service is then available simply by running
```
docker pull uhthomas/kipp
docker run uhthomas/kipp
```

## Support
Kipp is designed to be interoperable with a number of providers for both the
database, and the file system. The current support is limited, but it's trivial
to add new sources.

### Databases
* [Badger](https://github.com/dgraph-io/badger)

### File systems
File systems can be configured using the `--filesystem` flag. The flag requires
the input be parsable as a URL. See the [url.Parse](https://golang.org/pkg/net/url/#Parse)
docs for more info.
#### Local (your local file system)
The local filesystem does not require any special formatting, and can be used
like a regular path such

```
--filesystem /path/to/files
```

#### [AWS S3](https://aws.amazon.com/s3/)
AWS S3 requires the `s3` scheme, and has the following syntax:

```
--filesystem s3://some-token:some-secet@some-region/some-bucket?endpoint=some-endpoint.
```

The `region` and `bucket` are *required*.

The [user info](https://tools.ietf.org/html/rfc2396#section-3.2.2) section is
optional, if present, will create new static credentials. Otherwise, the default
AWS SDK credentials will be used.

The `endpoint` is optional, and will use the default AWS endpoint if not present.

#### [Backblaze B2](https://www.backblaze.com/b2/cloud-storage.html) (experimental)

## Building from source
Kipp is built, tested and compiled using [Bazel](https://bazel.build). To run
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