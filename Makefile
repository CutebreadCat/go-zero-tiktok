.PHONY: api-build gateway-build user-rpc video-rpc interaction-rpc communication-rpc \
        infra-pull infra-up infra-stop monitoring-up monitoring-stop \
        build-local run-gateway-local run-user-local run-video-local \
        run-interaction-local run-communication-local run-all-local \
        frontend-install frontend-dev frontend-build \
        test vet fmt api-get api-apifox db-shell mysql \
        migrate-up migrate-down \
        log-clean log-clean-dry log-clean-stop \
        log-cleaner-install log-cleaner-uninstall log-cleaner-status

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

# Infrastructure lifecycle (仅基础设施;监控见 monitoring-up)
infra-pull:
	docker compose -f deploy/docker-compose.yml pull

infra-up:
	docker compose -f deploy/docker-compose.yml up -d

infra-stop:
	docker compose -f deploy/docker-compose.yml stop

# Monitoring (Loki/Alloy/Grafana,可选)
monitoring-up:
	docker compose -f deploy/docker-compose.yml --profile monitoring up -d

monitoring-stop:
	docker compose -f deploy/docker-compose.yml --profile monitoring stop

# Prerequisite: start infrastructure on the host first, e.g. `make infra-up`.
# Local (non-docker) mode: binaries read app/*/etc/*.yaml, with host env vars
# pointing at 127.0.0.1 instead of docker service names. Override any var to
# target another environment (e.g. docker/k8s service names).
LOCAL_ENV := ETCD_HOSTS=127.0.0.1:2379 MYSQL_HOST=127.0.0.1 MYSQL_PORT=3309 MYSQL_PASSWORD=yourpassword REDIS_HOST=127.0.0.1:6888 KAFKA_BROKERS=127.0.0.1:9092 ACCESS_SECRET=your_access_secret OTLP_ENDPOINT=localhost:4317

build-local:
	go build -o bin/gateway ./app/gateway/api
	go build -o bin/user-rpc ./app/user/rpc
	go build -o bin/video-rpc ./app/video/rpc
	go build -o bin/interaction-rpc ./app/interaction/rpc
	go build -o bin/communication-rpc ./app/communication/rpc

run-gateway-local:
	$(LOCAL_ENV) ./bin/gateway -f app/gateway/api/etc/tiktok-api.yaml

run-user-local:
	$(LOCAL_ENV) ./bin/user-rpc -f app/user/rpc/etc/user.yaml

run-video-local:
	$(LOCAL_ENV) ./bin/video-rpc -f app/video/rpc/etc/video.yaml

run-interaction-local:
	$(LOCAL_ENV) ./bin/interaction-rpc -f app/interaction/rpc/etc/interaction.yaml

run-communication-local:
	$(LOCAL_ENV) ./bin/communication-rpc -f app/communication/rpc/etc/communication.yaml

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

# Frontend (Vue 3 + Vite)
frontend-install:
	cd frontend && npm ci

frontend-dev:
	cd frontend && npm install && npm run dev

frontend-build:
	cd frontend && npm ci && npm run build

# Documentation and local database access
# 使用 go-zero 内置 swagger 插件导出 .api 为 JSON（OpenAPI 2.0），可直接导入 Apifox
api-apifox:
	goctl api swagger --api api/main.api --dir docs --filename tiktok-api

api-get: api-apifox

db-shell:
	docker compose -f deploy/docker-compose.yml exec mysql mysql -uroot -pyourpassword

mysql: db-shell

# Database migrations (run against the docker mysql container only, never local MySQL)
MIGRATE_DSN := mysql://root:yourpassword@tcp(mysql:3306)/gozero-tiktok?charset=utf8mb4&parseTime=True&loc=Local
# 用 make 内置 $(CURDIR) 拼绝对路径，避免 Git-for-Windows/MSYS 把 ../migrations 解析成 Git 安装目录
MIGRATIONS_DIR := $(CURDIR)/migrations

migrate-up:
	MSYS_NO_PATHCONV=1 docker compose -f deploy/docker-compose.yml --profile migrate run --rm --no-deps \
		--volume "$(MIGRATIONS_DIR):/migrations" \
		--entrypoint migrate migrate \
		-path /migrations -database "$(MIGRATE_DSN)" up

migrate-down:
	MSYS_NO_PATHCONV=1 docker compose -f deploy/docker-compose.yml --profile migrate run --rm --no-deps \
		--volume "$(MIGRATIONS_DIR):/migrations" \
		--entrypoint migrate migrate \
		-path /migrations -database "$(MIGRATE_DSN)" down 1

# 日志清理(log-cleaner):过滤保留近 3 天日志,删除更早日志
# 默认守护轮询(1h);可用 LOG_ROOT/RETENTION/INTERVAL 环境变量覆盖
LOG_CLEANER := deploy/scripts/log-cleaner.sh

log-clean:            # 前台守护运行(按 Ctrl-C 停止)或配合 systemd
	bash $(LOG_CLEANER)

log-clean-dry:        # 试运行:只看会删/过滤什么,不真正删除
	DRY_RUN=1 LOG_ROOT=logs RETENTION=3 bash $(LOG_CLEANER)

log-clean-stop:       # 停止前台的守护进程(配合 log-clean 使用)
	@-pkill -f "$(LOG_CLEANER)" 2>/dev/null && echo "已停止 log-cleaner" || echo "log-cleaner 未在运行"

# 日志清理(log-cleaner)systemd 系统托管:
#   make log-cleaner-install [APP_DIR=/opt/tiktok]   # 安装 unit 并 enable --now(需要 sudo)
#   make log-cleaner-status                          # 查看服务状态
#   make log-cleaner-uninstall                       # 停止并移除 unit(需要 sudo)
# 注意:上面的 log-clean 是前台临时运行;systemd 托管后请用 systemctl 管理,
#       不要再混用 pkill 式的 log-clean-stop。
SYSTEMD_UNIT_DIR := /etc/systemd/system
LOG_CLEANER_UNIT := log-cleaner.service
LOG_CLEANER_UNIT_SRC := deploy/systemd/$(LOG_CLEANER_UNIT)
APP_DIR ?= /opt/tiktok

log-cleaner-install:
	sed 's|__APP_DIR__|$(APP_DIR)|g' $(LOG_CLEANER_UNIT_SRC) | sudo tee $(SYSTEMD_UNIT_DIR)/$(LOG_CLEANER_UNIT) >/dev/null
	sudo systemctl daemon-reload
	sudo systemctl enable --now log-cleaner
	@echo "log-cleaner 已安装并启动 (APP_DIR=$(APP_DIR));查看日志: journalctl -u log-cleaner -f"

log-cleaner-status:
	systemctl status log-cleaner --no-pager

log-cleaner-uninstall:
	@-sudo systemctl disable --now log-cleaner 2>/dev/null
	@-sudo rm -f $(SYSTEMD_UNIT_DIR)/$(LOG_CLEANER_UNIT)
	sudo systemctl daemon-reload
	@echo "log-cleaner 已卸载"
