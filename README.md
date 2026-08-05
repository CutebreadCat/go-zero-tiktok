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
- [快速开始](#快速开始)
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
- **互动服务**：视频点赞/取消、评论/回复/删除、评论点赞，幂等键保证重复请求安全
- **关系服务**：关注/取关、粉丝/关注/好友（互关）列表，`user_relation_stat` 原子计数，软删除可恢复
- **统一网关**：集中鉴权（`token.AuthMiddleware` + 公开路径白名单）、限流（`RateLimit`），RPC 间通过 etcd 服务发现
- **可观测性**：Prometheus 指标 + zap 结构化日志（自动注入 trace_id / span_id / user_id）
- **基础设施即代码**：Docker Compose 一键拉起 etcd / MySQL / Redis / Kafka / 迁移容器
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
        │ 账户 / MFA   │   │ 视频 / 搜索   │   │ 点赞 / 评论      │   │ 关注 / 好友      │
        └──────┬───────┘   └──────┬───────┘   └────────┬─────────┘   └────────┬────────┘
               └──────────────────┴──────────┬─────────┴──────────────────────┘
                                ┌────────────▼────────────┐
                                │  MySQL :3309  Redis :6888│
                                │  Kafka :9092  etcd :2379 │
                                └─────────────────────────┘
```

| 服务 | 说明 | 端口 | 目录 |
|---|---|---|---|
| gateway | HTTP 网关（鉴权、限流、聚合） | 8888 | `app/gateway/api` |
| user.rpc | 账户：注册、登录、MFA | 8890 | `app/user/rpc` |
| video.rpc | 视频：发布、搜索、热门、Feed | 8891 | `app/video/rpc` |
| interaction.rpc | 互动：点赞、评论、回复 | 8892 | `app/interaction/rpc` |
| communication.rpc | 关系：关注、粉丝、好友 | 8893 | `app/communication/rpc` |

各 RPC 服务内部采用 **domain / dal 分层**：`domain` 承载业务逻辑与仓储接口，`dal/reposity` 实现数据访问，`dal/tables` 定义 GORM 模型。

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
│   ├── kafka/  migrate/  storage/
├── Prometheus/logger/          # 结构化日志（链路字段注入）
├── migrations/                 # golang-migrate SQL 脚本
├── etc/                        # 本地运行配置（*-local.yaml）
├── testhelpers/                # 测试辅助（SQLite 建库、断言）
├── compose.infrastructure.yml  # 基础设施编排
├── Makefile                    # 一键命令入口
└── .github/workflows/lint.yml  # CI 静态检查
```

## 快速开始

### 前置条件

- Go 1.21+、Docker + Docker Compose
- `goctl`（代码生成，见 [代码生成](#代码生成)）

### 1. 启动基础设施

```bash
make infra-up
```

拉起 etcd、MySQL、Redis、Kafka 容器（MySQL 初始化库 `gozero-tiktok`，root 密码 `yourpassword`）。

### 2. 执行数据库迁移

```bash
make migrate-up
```

基于 `migrations/` 建表（迁移仅针对 Docker MySQL 容器运行）。

### 3. 构建

```bash
make build-local
```

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

> 本地模式下各服务读取 `etc/*-local.yaml`，将中间件指向宿主机（`127.0.0.1`）；如需容器化部署，修改对应服务配置中的连接地址为 Docker 服务名即可。

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
app/video/rpc/internal/dal/tables/video_baseinfo/           mysql_test.go
app/video/rpc/internal/dal/tables/video_liker/              mysql_test.go
```

```bash
make test
# 或
go test ./...
```

## 常用命令

| 命令 | 说明 |
|---|---|
| `make infra-up` / `make infra-stop` | 启动 / 停止基础设施容器 |
| `make migrate-up` / `make migrate-down` | 执行 / 回滚数据库迁移 |
| `make build-local` | 构建全部服务到 `bin/` |
| `make run-*-local` | 本地启动单个服务（gateway / user / video / interaction / communication） |
| `make test` / `make vet` / `make fmt` | 测试 / 静态检查 / 格式化 |
| `make db-shell` | 进入 MySQL 容器 |
| `make api-get` | 生成 Swagger 文档到 `docs/` |
| `make api-build` | goctl 重新生成网关代码 |

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
