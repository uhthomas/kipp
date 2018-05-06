FROM golang:alpine AS build
ADD . src/github.com/6f7262/kipp
RUN apk update && \
    apk add git build-base && \
    go get -v github.com/6f7262/kipp/cmd/kipp && \
    go build -v -ldflags "-linkmode external -extldflags -static" -o 
/kipp github.com/6f7262/kipp/cmd/kipp

FROM scratch
COPY --from=build /kipp /
ADD default /
CMD ["/kipp"]
