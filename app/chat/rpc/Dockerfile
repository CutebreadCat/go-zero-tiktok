FROM golang:1.26-alpine AS builder

ENV GOPROXY=https://mirrors.aliyun.com/goproxy/,https://goproxy.cn,https://mirrors.tencent.com/go/,https://mirrors.huaweicloud.com/repository/goproxy/,direct
ENV GOSUMDB=off
ENV GO111MODULE=on
ENV GOFLAGS="-buildvcs=false"

WORKDIR /app

RUN apk add --no-cache git

# 利用 BuildKit 缓存挂载，避免重复下载依赖
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o chat-rpc ./app/chat/rpc

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/chat-rpc ./chat-rpc
COPY --from=builder /app/app/chat/rpc/etc ./app/chat/rpc/etc
COPY --from=builder /app/pkg/storage/aliyun/aliconfig_example.yaml ./pkg/storage/aliyun/aliconfig.yaml

EXPOSE 8894

CMD ["./chat-rpc", "-f", "app/chat/rpc/etc/chat.yaml"]
