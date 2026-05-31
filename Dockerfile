FROM golang:1.26-alpine AS builder

ENV GOPROXY=https://mirrors.aliyun.com/goproxy/,https://goproxy.cn,https://mirrors.tencent.com/go/,https://mirrors.huaweicloud.com/repository/goproxy/,direct
ENV GOSUMDB=off
ENV GO111MODULE=on

WORKDIR /app

RUN apk add --no-cache git ffmpeg

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 运行测试


RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o tiktok .

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata ffmpeg

COPY --from=builder /app/tiktok ./tiktok
COPY --from=builder /app/etc ./etc
COPY --from=builder /app/internal/infra/storage/aliyun/aliconfig.yaml ./internal/infra/storage/aliyun/aliconfig.yaml

EXPOSE 8888

CMD ["./tiktok"]
