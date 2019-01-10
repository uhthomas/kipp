FROM golang:alpine AS build
RUN apk update && \
	apk add --no-cache git && \
	go get -d -v github.com/uhthomas/kipp/cmd/kipp && \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/src/github.com/uhthomas/kipp/default/kipp github.com/uhthomas/kipp/cmd/kipp

FROM scratch
COPY --from=build /go/src/github.com/uhthomas/kipp/default /
ENTRYPOINT ["/kipp", "--mime", "mime.json", "--files", "/data/files", "--tmp", "/data/tmp", "--store", "/data/kipp.db"]