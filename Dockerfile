FROM golang:1.14 as build

WORKDIR /build

COPY . .

RUN go build -ldflags "-s -w" -o /kipp ./cmd/kipp

FROM gcr.io/distroless/base-debian10

COPY --from=build /build/default /default

COPY --from=build /kipp /

ENTRYPOINT ["/kipp", "--mime", "mime.json", "--files", "/data/files", "--tmp", "/data/tmp", "--store", "/data/kipp.db"]