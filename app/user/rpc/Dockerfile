FROM golang:1.26-alpine AS builder

ENV GOPROXY=https://mirrors.aliyun.com/goproxy/,https://goproxy.cn,https://mirrors.tencent.com/go/,https://mirrors.huaweicloud.com/repository/goproxy/,direct
ENV GOSUMDB=off
ENV GO111MODULE=on

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o user-rpc ./app/user/rpc

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/user-rpc ./user-rpc
COPY --from=builder /app/app/user/rpc/etc ./app/user/rpc/etc

EXPOSE 8080

CMD ["./user-rpc", "-f", "app/user/rpc/etc/user.yaml"]
