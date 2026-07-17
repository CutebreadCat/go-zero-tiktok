.PHONY: api-build gateway-build user-rpc video-rpc interaction-rpc communication-rpc chat-rpc \
        infra-up infra-stop infra-down infra-restart infra-logs infra-ps \
        up stop down restart ps logs build \
        run-gateway run-user run-video run-interaction run-communication run-chat \
        stop-gateway stop-user stop-video stop-interaction stop-communication stop-chat \
        test vet fmt api-get db-shell mysql

# go-zero code generation
api-build:
	goctl api go -api api/main.api -dir app/gateway/api -style go_zero

user-rpc:
	goctl rpc protoc app/user/rpc/user.proto --go_out=app/user/rpc --go-grpc_out=app/user/rpc --zrpc_out=app/user/rpc --style go_zero

video-rpc:
	goctl rpc protoc app/video/rpc/video.proto --go_out=app/video/rpc --go-grpc_out=app/video/rpc --zrpc_out=app/video/rpc --style go_zero

interaction-rpc:
	goctl rpc protoc app/interaction/rpc/interaction.proto --go_out=app/interaction/rpc --go-grpc_out=app/interaction/rpc --zrpc_out=app/interaction/rpc --style go_zero

communication-rpc:
	goctl rpc protoc app/communication/rpc/communication.proto --go_out=app/communication/rpc --go-grpc_out=app/communication/rpc --zrpc_out=app/communication/rpc --style go_zero

chat-rpc:
	goctl rpc protoc app/chat/rpc/chat.proto --go_out=app/chat/rpc --go-grpc_out=app/chat/rpc --zrpc_out=app/chat/rpc --style go_zero

# Infrastructure lifecycle
infra-up:
	docker compose -f compose.infrastructure.yml up -d

infra-stop:
	docker compose -f compose.infrastructure.yml stop

infra-down:
	docker compose -f compose.infrastructure.yml down

infra-restart:
	docker compose -f compose.infrastructure.yml restart

infra-logs:
	docker compose -f compose.infrastructure.yml logs -f

infra-ps:
	docker compose -f compose.infrastructure.yml ps

# Microservice lifecycle
up:
	docker compose -f compose.infrastructure.yml -f compose.yml up -d --build

stop:
	docker compose -f compose.infrastructure.yml -f compose.yml stop

down:
	docker compose -f compose.infrastructure.yml -f compose.yml down

restart:
	docker compose -f compose.infrastructure.yml -f compose.yml restart

ps:
	docker compose -f compose.infrastructure.yml -f compose.yml ps

logs:
	docker compose -f compose.infrastructure.yml -f compose.yml logs -f

build:
	docker compose -f compose.infrastructure.yml -f compose.yml build

run-gateway:
	docker compose -f compose.infrastructure.yml -f compose.yml up -d --build gateway

run-user:
	docker compose -f compose.infrastructure.yml -f compose.yml up -d --build user-rpc

run-video:
	docker compose -f compose.infrastructure.yml -f compose.yml up -d --build video-rpc

run-interaction:
	docker compose -f compose.infrastructure.yml -f compose.yml up -d --build interaction-rpc

run-communication:
	docker compose -f compose.infrastructure.yml -f compose.yml up -d --build communication-rpc

run-chat:
	docker compose -f compose.infrastructure.yml -f compose.yml up -d --build chat-rpc

stop-gateway:
	docker compose -f compose.infrastructure.yml -f compose.yml stop gateway

stop-user:
	docker compose -f compose.infrastructure.yml -f compose.yml stop user-rpc

stop-video:
	docker compose -f compose.infrastructure.yml -f compose.yml stop video-rpc

stop-interaction:
	docker compose -f compose.infrastructure.yml -f compose.yml stop interaction-rpc

stop-communication:
	docker compose -f compose.infrastructure.yml -f compose.yml stop communication-rpc

stop-chat:
	docker compose -f compose.infrastructure.yml -f compose.yml stop chat-rpc

# Build and quality checks
gateway-build:
	go build ./app/gateway/api

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w app pkg

# Documentation and local database access
api-get:
	goctl api swagger --api api/main.api --dir docs

db-shell:
	docker compose -f compose.infrastructure.yml exec mysql mysql -uroot -pyourpassword

mysql: db-shell
