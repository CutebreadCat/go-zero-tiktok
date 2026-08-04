.PHONY: api-build gateway-build user-rpc video-rpc interaction-rpc communication-rpc \
        infra-up infra-stop \
        build-local run-gateway-local run-user-local run-video-local \
        run-interaction-local run-communication-local run-all-local \
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

# Infrastructure lifecycle
infra-up:
	docker compose -f compose.infrastructure.yml up -d

infra-stop:
	docker compose -f compose.infrastructure.yml stop

# Local (non-docker) mode: binaries read etc/*-local.yaml so that middleware
# endpoints point at the host (127.0.0.1) instead of docker service names.
# Prerequisite: start infrastructure on the host first, e.g. `make infra-up`.
LOCAL_ETC := $(CURDIR)/etc

build-local:
	go build -o bin/gateway ./app/gateway/api
	go build -o bin/user-rpc ./app/user/rpc
	go build -o bin/video-rpc ./app/video/rpc
	go build -o bin/interaction-rpc ./app/interaction/rpc
	go build -o bin/communication-rpc ./app/communication/rpc

run-gateway-local:
	./bin/gateway -f $(LOCAL_ETC)/gateway-local.yaml

run-user-local:
	./bin/user-rpc -f $(LOCAL_ETC)/user-local.yaml

run-video-local:
	./bin/video-rpc -f $(LOCAL_ETC)/video-local.yaml

run-interaction-local:
	./bin/interaction-rpc -f $(LOCAL_ETC)/interaction-local.yaml

run-communication-local:
	./bin/communication-rpc -f $(LOCAL_ETC)/communication-local.yaml

run-all-local: build-local
	@echo "Tip: run each service in its own terminal, e.g. 'make run-gateway-local'."

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
