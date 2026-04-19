FROM  golang:1.26-alpine

ENV GOPROXY=https://mirrors.aliyun.com/goproxy/,https://goproxy.cn,direct
ENV GOSUMDB=off
ENV GO111MODULE=on

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN cp config/config.yml ./config.yml

RUN go build -o main .
EXPOSE 8080

CMD ["./main"]