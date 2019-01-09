FROM golang:latest AS build
RUN go get -d -v github.com/uhthomas/kipp/cmd/kipp && \
    CGO_ENABLED=0 go build -o /go/src/github.com/uhthomas/kipp/default/kipp github.com/uhthomas/kipp/cmd/kipp

FROM scratch
COPY --from=build /go/src/github.com/uhthomas/kipp/default /
ENTRYPOINT ["/kipp", "--files", "/data/files", "--tmp", "/data/tmp", "--store", "/data/kipp.db"]