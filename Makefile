USER_RPC_PROTO := app/user/rpc/user.proto
USER_RPC_DIR := ./app/user/rpc

.PHONY : api-build mysql docker-build api-get coverage coverage-func coverage-html user-api user-rpc

api-build:
	goctl api go -api api/main.api -dir . -style go_zero

user-api:
	goctl api go -api app/user/api/user.api -dir ./app/user/api/ -style go_zero

user-rpc:
	goctl rpc protoc $(USER_RPC_PROTO) --go_out=$(USER_RPC_DIR) --go-grpc_out=$(USER_RPC_DIR) --zrpc_out=$(USER_RPC_DIR) --style go_zero

mysql:
    docker compose exec mysql mysql -uroot -pyourpassword

docker-build:
	docker-compose up --build

api-get:
	goctl api swagger --api api/main.api --dir docs

coverage:
	go test ./... -coverprofile coverage.out

coverage-func:
	go tool cover -func coverage.out

coverage-html:
	go tool cover -html coverage.out -o coverage.html
