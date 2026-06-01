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
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o user-rpc ./app/user/rpc

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/user-rpc ./user-rpc
COPY --from=builder /app/app/user/rpc/etc ./app/user/rpc/etc
COPY --from=builder /app/internal/infra/storage/aliyun/aliconfig.yaml ./internal/infra/storage/aliyun/aliconfig.yaml

EXPOSE 8080

CMD ["./user-rpc", "-f", "app/user/rpc/etc/user.yaml"]
