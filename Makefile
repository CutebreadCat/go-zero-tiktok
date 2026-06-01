USER_RPC_PROTO := app/user/rpc/user.proto
USER_RPC_DIR := ./app/user/rpc

.PHONY : api-build mysql docker-build api-get coverage coverage-func coverage-html user-api user-rpc
.PHONY : run-api run-user run-video run-interaction run-communication run-chat run-all

# Code generation
api-build:
	goctl api go -api api/main.api -dir . -style go_zero

user-api:
	goctl api go -api app/user/api/user.api -dir ./app/user/api/ -style go_zero

user-rpc:
	goctl rpc protoc $(USER_RPC_PROTO) --go_out=$(USER_RPC_DIR) --go-grpc_out=$(USER_RPC_DIR) --zrpc_out=$(USER_RPC_DIR) --style go_zero

# Run services
run-api:
	docker compose up -d --build go-zero-tiktokgo

run-user:
	docker compose up -d --build user-rpc

run-video:
	docker compose up -d --build video-rpc

run-interaction:
	docker compose up -d --build interaction-rpc

run-communication:
	docker compose up -d --build communication-rpc

run-chat:
	docker compose up -d --build chat-rpc

run-all:
	docker compose up -d --build

# Docker & Database
mysql:
    docker compose exec mysql mysql -uroot -pyourpassword

docker-build:
	docker-compose up --build

# Swagger docs
api-get:
	goctl api swagger --api api/main.api --dir docs

# Test coverage
coverage:
	go test ./... -coverprofile coverage.out

coverage-func:
	go tool cover -func coverage.out

coverage-html:
	go tool cover -html coverage.out -o coverage.html
