FROM golang:1.26-alpine AS builder

ENV GOPROXY=https://mirrors.aliyun.com/goproxy/,https://goproxy.cn,direct
ENV GOSUMDB=off
ENV GO111MODULE=on
ENV GOFLAGS="-buildvcs=false"

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gateway ./app/gateway/api

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/gateway ./gateway
COPY --from=builder /app/app/gateway/api/etc ./etc

EXPOSE 8888

CMD ["./gateway", "-f", "etc/tiktok-api.yaml"]
