# go-zero-tiktok

基于 [go-zero](https://go-zero.dev) 的短视频平台后端（仿 TikTok），采用 **网关 + 微服务** 架构，覆盖账户、视频、互动、关注关系四大业务域。

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![go-zero](https://img.shields.io/badge/go--zero-1.7-blue)](https://go-zero.dev)
[![gRPC](https://img.shields.io/badge/gRPC-Enabled-green)](https://grpc.io)
[![CI](https://img.shields.io/badge/CI-golangci--lint-blueviolet)](.github/workflows/lint.yml)
[![License](https://img.shields.io/badge/License-MIT-yellow)](LICENSE)

---

## 目录

- [功能特性](#功能特性)
- [系统架构](#系统架构)
- [技术栈](#技术栈)
- [目录结构](#目录结构)
- [快速开始（本地开发）](#快速开始本地开发)
- [服务器部署](#服务器部署)
- [API 一览](#api-一览)
- [错误码约定](#错误码约定)
- [测试](#测试)
- [常用命令](#常用命令)
- [代码生成](#代码生成)
- [相关文档](#相关文档)

---

## 功能特性

- **账户体系**：注册 / 登录 / 刷新令牌，JWT 双令牌（Access + Refresh），可选 MFA（多因素认证）二次校验
- **视频服务**：发布视频、作者作品列表、关键词搜索、热门排行、Feed 信息流（游标/分页）
- **互动服务**：视频点赞/取消、收藏/取消收藏、评论/回复/删除、评论点赞；点赞/收藏已走“Redis + Kafka + 兜底 Syncer”异步链路
- **关系服务**：关注/取关、粉丝/关注/好友（互关）列表，`user_relation_stat` 原子计数，软删除可恢复
- **统一网关**：集中鉴权（`token.AuthMiddleware` + 公开路径白名单）、限流（`RateLimit`），RPC 间通过 etcd 服务发现
- **可观测性**：Prometheus 指标 + zap 结构化日志（自动注入 trace_id / span_id / user_id）
- **基础设施即代码**：Docker Compose 一键拉起 etcd / MySQL / Redis / Kafka；业务服务直接以二进制运行
- **质量保障**：表驱动单元测试（纯 Go SQLite，无 cgo）、golangci-lint CI

## 系统架构

```
                          ┌─────────────────────┐
                          │   HTTP 网关 :8888    │
                          │    app/gateway/api   │
                          │  鉴权 · 限流 · 路由   │
                          └───┬────┬────┬────┬───┘
                              │    │    │    │      gRPC（etcd 服务发现 :2379）
                ┌─────────────┘    │    │    └─────────────┐
        ┌───────▼──────┐   ┌───────▼──────┐   ┌────────────▼─────┐   ┌─────────────────┐
        │  user.rpc    │   │  video.rpc   │   │ interaction.rpc  │   │ communication.rpc│
        │     :8890    │   │    :8891     │   │      :8892       │   │      :8893      │
        │ 账户 / MFA   │   │ 视频 / 搜索   │   │ 点赞/收藏/评论   │   │ 关注 / 好友      │
        └──────┬───────┘   └──────┬───────┘   └────────┬─────────┘   └────────┬────────┘
               └──────────────────┴──────────┬─────────┴──────────────────────┘
                                ┌────────────▼────────────┐
                                │  MySQL :3309  Redis :6888│
                                │  Kafka :9092  etcd :2379 │
                                └─────────────────────────┘
```

| 服务 | 说明 | 端口 | 指标端口 | 目录 |
|---|---|---|---|---|
| gateway | HTTP 网关（鉴权、限流、聚合） | 8888 | 9100 | `app/gateway/api` |
| user.rpc | 账户：注册、登录、MFA | 8890 | 9101 | `app/user/rpc` |
| video.rpc | 视频：发布、搜索、热门、Feed | 8891 | 9102 | `app/video/rpc` |
| interaction.rpc | 互动：点赞、收藏、评论、回复 | 8892 | 9103 | `app/interaction/rpc` |
| communication.rpc | 关系：关注、粉丝、好友 | 8893 | 9104 | `app/communication/rpc` |

各 RPC 服务内部采用 **domain / dal 分层**：`domain` 承载业务逻辑与仓储接口，`dal/reposity` 实现数据访问，`dal/tables` 定义 GORM 模型。

> **组织约定**：业务领域统一收敛到 `internal/domain`（可建子目录），不在 `internal/` 根级别新增与 `domain` 平层的目录；RPC 服务之间禁止相互调用，跨服务数据需求由网关编排或共享数据库表满足。

## 技术栈

| 分类 | 选型 |
|---|---|
| 语言 / 框架 | Go 1.21+、go-zero（goctl 代码生成） |
| ORM | GORM |
| 服务通信 | gRPC + protobuf、etcd 服务发现 |
| 存储 | MySQL 8.0、Redis 7.2（Kafka 4.0 预留消息能力） |
| 鉴权 | JWT（Access + Refresh）、MFA（TOTP 二维码） |
| 分布式 ID | snowflake（`pkg/utils`） |
| 日志 / 监控 | zap 结构化日志、Prometheus 指标、trace/span 链路注入 |
| 测试 | 标准库 `testing`（表驱动）+ glebarez/sqlite（纯 Go 内存库） |
| 工程化 | golangci-lint、GitHub Actions、Docker Compose、golang-migrate |

## 目录结构

```
go-zero-tiktok/
├── api/                        # goctl API 契约（接口唯一事实来源）
│   ├── main.api                #   聚合入口（import 各子模块）
│   ├── model.api               #   共享模型（BaseResponse、UserBaseinfo …）
│   ├── user/  video/  interaction/  communication/
│   └── └─ *.api + *_auth.api   #   公开接口与需登录接口分离定义
├── app/
│   ├── gateway/api/            # HTTP 网关（goctl 生成 + 手写中间件）
│   │   ├── etc/                #   服务配置（tiktok-api.yaml，环境变量注入）
│   │   └── internal/
│   │       ├── handler/ logic/ #   路由处理与业务逻辑
│   │       ├── middleware/     #   token 鉴权、RateLimit 限流
│   │       ├── svc/ types/     #   依赖注入上下文、API 类型
│   ├── user/rpc/               # 用户服务（proto + zrpc + domain/dal）
│   ├── video/rpc/              # 视频服务
│   ├── interaction/rpc/        # 互动服务
│   └── communication/rpc/      # 关系服务
├── pkg/                        # 共享库
│   ├── contract/               #   RPC 共享契约（DTO + context 键）
│   ├── jwt/  xerr/  utils/     #   JWT、错误码、工具（ID/密码）
│   ├── kafka/  storage/
│   └── logger/                 #   结构化日志（链路字段注入）
├── migrations/                 # golang-migrate SQL 脚本
├── testhelpers/                # 测试辅助（SQLite 建库、断言）
├── deploy/
│   ├── docker-compose.yml      # 基础设施（etcd/redis/mysql/kafka）+ 迁移 + 监控编排
│   ├── monitoring/             #   可选：Loki / Alloy / Grafana 监控配置
│   └── log-cleaner/            #   日志定期清理脚本
├── Makefile                    # 一键命令入口
└── .github/workflows/lint.yml  # CI 静态检查
```

## 快速开始（本地开发）

### 前置条件

- Go 1.21+、Docker + Docker Compose
- `goctl`（代码生成，见 [代码生成](#代码生成)）

### 1. 启动基础设施

```bash
make infra-pull   # 预拉取 compose 中全部镜像(etcd/MySQL/Redis/Kafka/migrate/监控,首次部署建议执行)
make infra-up
```

拉起 etcd、MySQL、Redis、Kafka 容器（MySQL 初始化库 `gozero-tiktok`，root 密码 `yourpassword`，详见下文「修改默认密码」）。

### 2. 执行数据库迁移

```bash
make migrate-up
```

基于 `migrations/` 增量建表（迁移仅针对 Docker MySQL 容器运行）。

### 3. 构建

```bash
make build-local
```

产物输出到 `bin/`（gateway / user-rpc / video-rpc / interaction-rpc / communication-rpc）。

### 4. 启动服务

建议每个服务开一个终端：

```bash
make run-gateway-local       # HTTP :8888
make run-user-local          # RPC  :8890
make run-video-local         # RPC  :8891
make run-interaction-local   # RPC  :8892
make run-communication-local # RPC  :8893
```

### 5. 验证

```bash
# 注册 → 登录 → 携带令牌访问需登录接口
curl -X POST http://localhost:8888/users \
  -d "username=alice&password=123456"
```

> 配置全部通过**环境变量**注入（见下文「环境变量说明」），`Makefile` 的 `LOCAL_ENV` 已默认指向本机 `127.0.0.1`；部署到服务器时只需改用服务器地址即可，服务配置（`app/*/etc/*.yaml`）无需修改。

## 服务器部署

> 部署形态：**基础设施（etcd/MySQL/Redis/Kafka）跑 Docker 容器，5 个业务服务直接编译为二进制运行**，由 systemd 托管。所有连接地址、密钥通过环境变量注入。

### 环境变量说明

服务配置文件（`app/*/etc/*.yaml`）中的连接信息均为 `${ENV}` 占位符，运行时由环境变量注入，**部署时无需改任何 yaml**：

| 变量 | 用途 | 本地默认值 |
|---|---|---|
| `ETCD_HOSTS` | etcd 地址（服务发现） | `127.0.0.1:2379` |
| `MYSQL_HOST` | MySQL 地址 | `127.0.0.1` |
| `MYSQL_PORT` | MySQL 端口（宿主机映射） | `3309` |
| `MYSQL_PASSWORD` | MySQL 密码 | `yourpassword` |
| `REDIS_HOST` | Redis 地址 | `127.0.0.1:6888` |
| `KAFKA_BROKERS` | Kafka 地址（视频 like 事件异步链路） | `127.0.0.1:9092` |
| `ACCESS_SECRET` | JWT 签名密钥（**5 个服务必须一致**） | `your_access_secret` |
| `OTLP_ENDPOINT` | 链路追踪导出地址（可选，不开监控可留空） | `localhost:4317` |

### 部署步骤（以 Ubuntu 22.04 为例）

#### 1. 安装依赖

```bash
# Docker + Compose 插件（基础设施用）
sudo apt update && sudo apt install -y docker.io docker-compose-plugin
sudo systemctl enable --now docker
sudo usermod -aG docker $USER && newgrp docker   # 免 sudo 使用 docker

# Go 1.21+（业务服务编译用）
wget https://go.dev/dl/go1.22.12.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.22.12.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc && source ~/.bashrc
```

#### 2. 拉取代码并构建

```bash
git clone <你的仓库地址> go-zero-tiktok && cd go-zero-tiktok
make build-local            # 编译 5 个二进制到 bin/
```

#### 3. 修改默认密码与密钥（生产环境必须）

MySQL 密码与 JWT 密钥默认是硬编码的，上线前必须修改。集中写入环境文件，避免散落各处：

```bash
# 生成随机密钥
openssl rand -hex 32    # 复制输出作为 ACCESS_SECRET
sudo mkdir -p /etc/go-zero-tiktok
cat <<'EOF' | sudo tee /etc/go-zero-tiktok/env
ETCD_HOSTS=127.0.0.1:2379
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3309
MYSQL_PASSWORD=你的新密码
REDIS_HOST=127.0.0.1:6888
KAFKA_BROKERS=127.0.0.1:9092
ACCESS_SECRET=你的随机密钥
OTLP_ENDPOINT=localhost:4317
EOF
sudo chmod 600 /etc/go-zero-tiktok/env
```

> ⚠️ **修改 MySQL 密码的联动点**（三处必须一致，否则服务连不上库 / 迁移失败）：
> 1. 上面的环境文件 `MYSQL_PASSWORD`
> 2. `deploy/docker-compose.yml` 中 MySQL 的 `MYSQL_PASSWORD`（或在该文件同级建 `.env` 覆盖，见步骤 4）
> 3. `Makefile` 中 `MIGRATE_DSN` 里的密码（迁移工具使用）
>
> 注意：MySQL 容器**首次启动时**才会用你设置的密码初始化数据卷；若已用旧密码启动过，需 `docker compose -f deploy/docker-compose.yml down`（保留数据卷）后删除 `mysql-data` 卷再重建，或手动 `ALTER USER` 改密。

#### 4. 启动基础设施并迁移

```bash
# 在 deploy/ 下建 .env 覆盖默认密码（不建则用默认 yourpassword）
cat > deploy/.env <<'EOF'
MYSQL_PASSWORD=你的新密码
EOF

make infra-pull    # 预拉取基础设施镜像(可跳过)
make infra-up      # 启动 etcd / MySQL / Redis / Kafka
make migrate-up    # 建表（增量，可重复执行）
```

#### 5. 用 systemd 托管 5 个服务（推荐）

创建 5 个 unit 文件（以 gateway 为例）：

```ini
# /etc/systemd/system/gozero-gateway.service
[Unit]
Description=go-zero-tiktok gateway
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
WorkingDirectory=/opt/go-zero-tiktok
EnvironmentFile=/etc/go-zero-tiktok/env
ExecStart=/opt/go-zero-tiktok/bin/gateway -f /opt/go-zero-tiktok/app/gateway/api/etc/tiktok-api.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

其余 4 个仅 `ExecStart` 不同：

| 服务 | ExecStart |
|---|---|
| user-rpc | `bin/user-rpc -f app/user/rpc/etc/user.yaml` |
| video-rpc | `bin/video-rpc -f app/video/rpc/etc/video.yaml` |
| interaction-rpc | `bin/interaction-rpc -f app/interaction/rpc/etc/interaction.yaml` |
| communication-rpc | `bin/communication-rpc -f app/communication/rpc/etc/communication.yaml` |

> 若代码不在 `/opt/go-zero-tiktok`，把 `WorkingDirectory` 和 `ExecStart` 路径改为实际路径即可。

```bash
sudo systemctl daemon-reload
# 启动顺序：先 RPC，后 gateway
sudo systemctl enable --now gozero-user-rpc gozero-video-rpc gozero-interaction-rpc gozero-communication-rpc gozero-gateway
sudo systemctl status gozero-gateway
```

#### 6. 验证

```bash
curl -X POST http://localhost:8888/users -d "username=alice&password=123456"
# 期望：{"status_code":0,"status_msg":"ok",...}
```

### 防火墙与安全

| 端口 | 服务 | 建议 |
|---|---|---|
| 8888 | gateway HTTP（对外） | 必须放行 |
| 2379 / 3309 / 6888 / 9092 | etcd / MySQL / Redis / Kafka | **禁止对外**，仅内网/本机 |
| 9100-9104 | Prometheus 指标 | 按需内网放行 |
| 3000 | Grafana（若启用监控） | 内网或 VPN 访问 |
| 4317 | OTLP 链路（若启用） | 仅本机 |

```bash
sudo ufw allow 8888/tcp        # 只对公网开网关
```

### 日志与清理

- 业务日志：`logs/{service}/{date}/{service}.log`（按天分目录，自动注入 trace_id）
- 查看 systemd 服务日志：`journalctl -u gozero-gateway -f`
- 定期清理日志（默认保留 3 天，每小时巡检）：

```bash
sudo tee /etc/systemd/system/gozero-log-cleaner.service >/dev/null <<'EOF'
[Unit]
Description=go-zero-tiktok log cleaner

[Service]
Type=simple
WorkingDirectory=/opt/go-zero-tiktok
ExecStart=/bin/bash /opt/go-zero-tiktok/deploy/log-cleaner/log-cleaner.sh
Restart=always
EOF
sudo systemctl enable --now gozero-log-cleaner
```

### 可选：启用监控（Loki / Alloy / Grafana）

```bash
make monitoring-up
# 浏览器访问 http://<服务器IP>:3000  (admin / admin)
# Grafana → Explore → 数据源 Loki，用 LogQL 查日志，如 {service="user-rpc"}
```

### 更新与回滚

```bash
# 拉新代码 → 若 migrations/ 有新增版本则先迁移 → 重编译 → 重启
git pull
make migrate-up          # 增量应用新迁移
make build-local
sudo systemctl restart gozero-user-rpc gozero-video-rpc gozero-interaction-rpc gozero-communication-rpc gozero-gateway
```

回滚迁移：`make migrate-down`（回滚最近 1 个版本），然后重启受影响服务。

### 常见问题

| 现象 | 排查 |
|---|---|
| 服务起不来，报连不上数据库 | 检查 `MYSQL_HOST/PORT/PASSWORD` 三处是否一致；`docker compose -f deploy/docker-compose.yml ps` 看 mysql 是否 running |
| gateway 报 RPC 服务不可用 | etcd 没起来或 `ETCD_HOSTS` 不对；`make infra-up` 后确认 etcd 容器 running |
| 迁移报 `no change` | 正常，说明已是最新版本（幂等） |
| 改了密码后迁移失败 | `MIGRATE_DSN` 里的密码没同步改（见步骤 3 联动点） |
| 迁移报 dirty | 上次迁移中断，需 `make migrate-down` 回滚后重试 |
| 忘记 ACCESS_SECRET 不一致 | 5 个服务必须用同一个值，否则 token 校验失败（401） |

## API 一览

### 用户 `user`

| 方法 | 路径 | 说明 | 鉴权 |
|---|---|---|---|
| POST | `/users` | 注册 | 公开 |
| POST | `/sessions` | 登录 | 公开 |
| POST | `/sessions/refresh` | 刷新访问令牌 | 公开 |
| GET | `/users/me` | 当前用户信息 | 需登录 |
| PUT | `/users/me/photo` | 更新头像 | 需登录 |
| GET | `/users/me/mfa/qr` | 获取 MFA 绑定二维码 | 需登录 |
| POST | `/users/me/mfa/bind` | 绑定 MFA | 需登录 |

### 视频 `video`

| 方法 | 路径 | 说明 | 鉴权 |
|---|---|---|---|
| POST | `/videos` | 发布视频 | 需登录 |
| GET | `/users/:id/videos` | 作者视频列表 | 公开 |
| GET | `/videos/search` | 关键词搜索 | 公开 |
| GET | `/videos/popular` | 热门视频 | 公开 |
| GET | `/feed-items` | Feed 信息流 | 公开 |

### 互动 `interaction`

| 方法 | 路径 | 说明 | 鉴权 |
|---|---|---|---|
| PUT | `/videos/:id/like` | 点赞视频 | 需登录 |
| DELETE | `/videos/:id/like` | 取消点赞 | 需登录 |
| GET | `/users/me/likes` | 我的点赞列表 | 需登录 |
| PUT | `/videos/:id/favorite` | 收藏视频 | 需登录 |
| DELETE | `/videos/:id/favorite` | 取消收藏 | 需登录 |
| GET | `/users/me/favorites` | 我的收藏列表 | 需登录 |
| POST | `/videos/:id/comments` | 发表评论 | 需登录 |
| GET | `/videos/:id/comments` | 评论列表 | 需登录 |
| DELETE | `/comments/:id` | 删除评论 | 需登录 |
| PUT | `/comments/:id/like` | 点赞评论 | 需登录 |
| DELETE | `/comments/:id/like` | 取消点赞评论 | 需登录 |
| POST | `/comments/:id/replies` | 回复评论 | 需登录 |

### 关系 `communication`

| 方法 | 路径 | 说明 | 鉴权 |
|---|---|---|---|
| PUT | `/users/me/following/:id` | 关注用户 | 需登录 |
| DELETE | `/users/me/following/:id` | 取消关注 | 需登录 |
| GET | `/users/me/following` | 我的关注列表 | 需登录 |
| GET | `/users/me/followers` | 我的粉丝列表 | 需登录 |
| GET | `/users/me/friends` | 我的好友（互关）列表 | 需登录 |

## 错误码约定

统一响应结构 `BaseResponse`：

```json
{ "status_code": 1002, "status_msg": "参数错误" }
```

| 错误码 | 含义 |
|---|---|
| 1001 | 系统繁忙 / 服务端异常 |
| 1002 | 参数错误（非法入参） |
| 401 | 未授权 / 令牌无效或过期 |

## 测试

基于标准库 `testing` 的**表驱动**测试，覆盖核心表的仓储层（成功 / 参数错误 / 鉴权 / 幂等 / 游标分页路径），使用纯 Go SQLite（glebarez/sqlite）保证可移植，无需外部数据库：

```
app/communication/rpc/internal/dal/tables/user_follow/      mysql_test.go
app/communication/rpc/internal/dal/tables/user_relation_stat/mysql_test.go
app/interaction/rpc/internal/dal/tables/comment_baseinfo/   mysql_test.go
app/interaction/rpc/internal/dal/tables/video_interaction/  mysql_test.go
app/video/rpc/internal/dal/tables/video_baseinfo/           mysql_test.go
```

```bash
make test
# 或
go test ./...
```

## 常用命令

| 命令 | 说明 |
|---|---|
| `make infra-pull` | 预拉取 compose 中全部镜像（含监控，避免逐个下载） |
| `make infra-up` / `make infra-stop` | 启动 / 停止基础设施容器 |
| `make migrate-up` / `make migrate-down` | 执行 / 回滚数据库迁移 |
| `make build-local` | 构建全部服务到 `bin/` |
| `make run-*-local` | 前台启动单个服务（gateway / user / video / interaction / communication） |
| `make monitoring-up` / `make monitoring-stop` | 启停监控（Loki/Alloy/Grafana，可选） |
| `make test` / `make vet` / `make fmt` | 测试 / 静态检查 / 格式化 |
| `make db-shell` | 进入 MySQL 容器 |
| `make api-get` | 生成 Swagger 文档到 `docs/` |
| `make api-build` | goctl 重新生成网关代码 |
| `make log-clean` | 前台运行日志清理守护（默认保留 3 天） |

## 代码生成

接口契约（`.api`）与 RPC 定义（`.proto`）为唯一事实来源，使用 goctl 生成骨架代码：

```bash
# HTTP 网关（依据 api/main.api）
make api-build

# RPC 服务（依据 app/<service>/rpc/*.proto）
make user-rpc video-rpc interaction-rpc communication-rpc
```

> 生成物带有 `// Code generated by goctl. DO NOT EDIT.` 标记，业务代码（domain / dal / middleware）为手工维护。

## 相关文档

- [目录结构说明](docs/directory.md)
- [API 接口文档](docs/api.md)
- [Kafka 消息消费架构](docs/mq.md)
