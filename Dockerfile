FROM golang:1.14 as build

WORKDIR /build

COPY . .

RUN go build -ldflags "-s -w" -o /kipp ./cmd/kipp

FROM gcr.io/distroless/base-debian10

COPY --from=build /build/default /default

COPY --from=build /kipp /

ENTRYPOINT ["/kipp", "--dir", "/data/files", "--tmp", "/data/tmp", "--dsn", "/data/kipp.db"]