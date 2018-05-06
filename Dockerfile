FROM golang:alpine AS build
ADD . src/github.com/6f7262/kipp
RUN apk update && \
    apk add git build-base && \
    go get -v github.com/6f7262/kipp/cmd/kipp && \
    go build -v -o /kipp github.com/6f7262/kipp/cmd/kipp

FROM alpine
COPY --from=build /kipp /
ADD default /
RUN mkdir -p /data/files && \
    ln -s /data/files /files && \
    ln -s /data/kipp.db /kipp.db
CMD ["/kipp"]
