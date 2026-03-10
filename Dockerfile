FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY main.go .

RUN go mod init proxy \
 && go get github.com/armon/go-socks5 \
 && go get github.com/elazarl/goproxy \
 && go build -o proxy

FROM alpine:3.20

COPY --from=builder /app/proxy /usr/local/bin/proxy

EXPOSE 14050 14051

CMD ["/usr/local/bin/proxy"]