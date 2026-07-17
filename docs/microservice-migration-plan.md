# go_zero-tiktok 微服务迁移计划书

## 1. 文档目的

本文用于指导 `go_zero-tiktok` 从当前“模块化单体 + 部分 RPC 服务”迁移为“独立 API 网关 + 领域 RPC 服务 + 可复用公共包”的架构。

本次重构的第一阶段只解决代码边界、启动边界和依赖方向，不追求一次性完成物理数据库拆分。迁移完成后，业务能力可以在同一 monorepo 中独立构建、独立部署，并为后续拆分数据库、消息服务和独立仓库做好准备。

## 2. 当前状态评估

当前项目已经具备以下基础：

- `app/user/rpc`、`app/video/rpc`、`app/interaction/rpc`、`app/communication/rpc`、`app/chat/rpc` 已存在，并且已经采用 go-zero RPC 入口。
- 根目录 `tiktok.go` 仍然启动唯一 HTTP 服务，路由定义位于根级 `api`，处理器、业务逻辑和依赖注入位于根级 `internal`。
- 根级 `internal` 同时承担了 API 网关、领域业务、基础设施、公共类型和公共工具等多个职责。
- RPC 服务仍然直接依赖 `go_zero-tiktok/internal/...`，因此当前服务之间存在源码级耦合，尚未达到独立服务边界。
- `compose.yml` 已经同时编排 API 单体、多个 RPC、etcd、MySQL、Redis 和 Kafka，但 API 单体仍直接依赖数据库和基础设施。

当前主要问题：

1. HTTP 网关和业务实现处于同一进程，网关无法独立扩缩容。
2. `internal` 目录边界过大，公共能力、领域能力和基础设施相互引用。
3. API 逻辑可以直接触达 `svc`、`dal` 或基础设施，跨领域依赖难以控制。
4. 多个服务共享数据库和表访问代码，后续拆分时容易形成分布式事务和数据所有权冲突。
5. 根级配置、Dockerfile 和启动入口无法按服务独立演进。

## 3. 迁移目标

### 3.1 目标架构

```text
客户端
  -> app/gateway/api（唯一外部 HTTP/WS 入口）
       -> user-rpc
       -> video-rpc
       -> interaction-rpc
       -> communication-rpc（后续可更名 relation-rpc）
       -> chat-rpc / chat-ws
       -> 后续 ai-rpc

各领域 RPC
  -> 本领域 application/domain
  -> 本领域 repository/model
  -> 本领域数据表、缓存和消息

公共包 pkg
  -> 只提供稳定、无业务归属的通用能力和契约
```

### 3.2 目标职责

| 组件 | 负责内容 | 明确禁止 |
| --- | --- | --- |
| `app/gateway/api` | HTTP/WS 协议、参数校验、JWT 校验、限流、响应格式、RPC 编排 | 直接访问数据库、直接引用领域 DAL、实现核心业务规则 |
| `app/<domain>/rpc` | 本领域业务用例、领域规则、数据读写、领域事件 | 直接读取其他领域的表，暴露本地 repository 给其他服务 |
| `pkg` | 错误码、上下文键、JWT 契约、ID、消息契约、纯工具 | 保存业务状态、持有数据库连接、依赖具体领域 `svc` |
| `internal` | 迁移期间的兼容代码，最终仅保留服务私有实现 | 被新服务当作公共依赖长期引用 |

## 4. 总体迁移策略

采用“先外置网关、再公共包迁移、最后领域收敛”的渐进式路线。每个阶段都必须能够编译、测试和回滚，避免一次性大范围移动导致无法定位问题。

原则如下：

- 保持对外 HTTP 路径和响应格式兼容，优先做到客户端无感。
- 先改变依赖方向，再改变物理目录；先建立接口，再移动实现。
- 迁移期间允许 `internal` 和 `pkg` 并存，但禁止新增业务代码依赖根级 `internal`。
- 业务跨域调用只允许通过 RPC client、事件或明确的接口适配器完成。
- 每个服务的 `svc` 只注入本服务需要的 repository、domain service、RPC client 和基础设施适配器，不注入无关中间件实例。
- 第一阶段可以继续共用一个 MySQL 实例，但代码上必须明确表的归属，禁止跨服务直接读写他域表。

## 5. 目标目录结构

第一阶段建议保持 monorepo，新增和调整目录如下：

```text
go_zero-tiktok/
  app/
    gateway/
      api/
        etc/gateway.yaml
        internal/config/
        internal/handler/
        internal/logic/
        internal/svc/
        gateway.go
      Dockerfile
    user/rpc/
    video/rpc/
    interaction/rpc/
    communication/rpc/
    chat/rpc/
    chat/ws/                 # 第二阶段按需要拆出
  pkg/
    ctxkey/
    xerr/
    jwt/
    id/
    mq/
    utils/
    contract/                # 对外稳定请求/响应或事件契约
  api/
    gateway/                  # 网关 API 定义迁移后的规范来源
  deploy/
    docker/
    k8s/                      # 后续增加
  docs/
  sql/
```

说明：`app/<domain>/rpc/internal` 是服务私有目录，可以保留；本次要求的 `internal` 到 `pkg` 迁移，针对根级可复用代码，不是把所有服务私有实现都暴露为公共包。

## 6. 分阶段实施计划

### 阶段 0：基线和冻结规则（1 周）

目标是建立可比较、可回滚的迁移基线。

工作项：

1. 固化当前接口清单：用户、视频、互动、关系、聊天和 WebSocket 路径。
2. 为现有 API、RPC 和关键业务补齐 smoke test、接口回归测试和构建检查。
3. 记录当前数据库表、Redis key、Kafka topic、外部 OSS/AI 依赖及其所有者。
4. 建立架构检查规则：新代码禁止新增 `go_zero-tiktok/internal` 的跨服务引用。
5. 保留当前 `tiktok.go` 作为迁移期间的回滚入口，不立即删除。

交付物：接口基线、依赖清单、数据归属表、回滚分支或可部署镜像。

### 阶段 1：外置 API 网关（1～2 周）

这是本次迁移的第一优先级。

#### 1.1 新建网关服务

- 创建 `app/gateway/api`，从根级 `api` 迁移或重新整理 `.api` 定义。
- 使用 go-zero 生成网关的 handler、logic、types 和路由注册代码。
- 将根级 `tiktok.go` 的启动逻辑迁移为 `app/gateway/api/gateway.go`。
- 新建 `etc/gateway.yaml`，仅保留 Rest、JWT、限流、RPC client、超时和链路追踪配置。
- 网关的 `ServiceContext` 只注入 RPC client、鉴权组件、限流组件、响应转换器和必要的轻量配置。

#### 1.2 网关逻辑改造

当前 `internal/logic/*` 中属于 HTTP 用例编排的代码迁移到 `app/gateway/api/internal/logic`，但逻辑必须改为：

```text
请求解析/校验 -> 提取用户身份 -> 调用 RPC client -> 聚合结果 -> 输出 HTTP 响应
```

禁止网关逻辑：

- 调用 GORM、SQL、Redis、Kafka、OSS 或 AI SDK。
- 直接 import `app/<domain>/rpc/internal/dal`。
- 通过共享 Go struct 绕过 RPC 契约访问服务内部实现。
- 在网关中实现注册、发布视频、评论、关注等领域规则。

#### 1.3 路由迁移顺序

建议按以下顺序接入 RPC，降低一次性风险：

1. 用户：注册、登录、刷新 token、用户信息。
2. 视频：列表、搜索、热门、Feed、发布。
3. 互动：点赞、评论、评论回复和删除。
4. 关系：关注、粉丝、好友列表。
5. 聊天：房间、消息、加入房间。
6. WebSocket：先保持现有协议，后续再独立 `chat/ws`。

每迁移一个领域，保留旧路径的兼容行为，并通过配置开关在旧单体和新网关之间切换。

#### 1.4 网关验收标准

- `app/gateway/api` 可以独立编译和启动。
- 网关容器不需要 MySQL、Redis、Kafka、OSS 配置即可启动；这些依赖只由下游服务持有。
- 所有既有 HTTP 路径通过回归测试，响应结构和错误码保持兼容。
- 网关进程停止不会导致 RPC 服务停止，网关可以单独水平扩容。
- `compose.yml` 中以 `gateway` 替换根级 `go-zero-tiktokgo` 对外暴露 8888 端口。

### 阶段 2：根级 `internal` 迁移为 `pkg`（1～2 周）

本阶段只迁移真正通用的代码，避免把领域实现包装成“公共包”。

#### 2.1 推荐迁移映射

| 当前路径 | 目标路径 | 迁移条件 |
| --- | --- | --- |
| `internal/shared/ctxkey` | `pkg/ctxkey` | 只包含上下文 key 和取值约定 |
| `internal/shared/xerr` | `pkg/xerr` | 错误码、错误包装、统一错误响应 |
| `internal/shared/mq` | `pkg/mq` | 消息信封、事件类型、路由契约，不放 Kafka 实现 |
| `internal/middleware/token` | `pkg/jwt` | JWT claims、token 校验契约；HTTP middleware 留在网关 |
| `internal/utils/id.go` | `pkg/id` | ID 生成及解析，不能依赖领域 svc |
| `internal/utils/string.go`、`time.go`、`encryption.go` | `pkg/utils` | 纯函数、无数据库和业务依赖 |
| `internal/types` | 按使用场景拆到 `pkg/contract` 或各服务内部 | 不保留大而全的共享领域模型 |
| `internal/infra/*` | 暂不整体迁移 | 具体实现必须归属使用它的服务，或后续单独抽象接口 |
| `internal/svc`、`internal/handler`、`internal/logic` | 不迁移到 `pkg` | 它们是应用或服务私有代码 |

#### 2.2 迁移规则

1. 先复制并改 import，验证所有服务通过测试，再删除旧包。
2. `pkg` 中的包必须有稳定的公开 API、单元测试和使用说明。
3. `pkg` 不允许反向依赖 `app/*` 或根级 `internal`。
4. 领域 DTO 不直接共享；跨服务只共享 proto、事件 schema 或最小化 contract。
5. 迁移期间使用兼容别名或适配器，避免一次性修改全部调用方。

#### 2.3 完成标志

```text
rg 'go_zero-tiktok/internal/' app pkg api
```

结果只允许出现服务私有的兼容引用；新建代码不得再引入根级 `internal`。根级 `internal` 在没有引用后删除，或仅保留迁移说明文件。

### 阶段 3：服务内部高内聚和依赖注入整改（2～3 周）

本阶段落实“logic 不直接引用 dal”和“svc 高内聚”的要求。

#### 3.1 推荐服务内部分层

```text
app/<domain>/rpc/internal/
  config/          # 配置结构
  server/          # RPC transport 适配
  logic/           # RPC 请求编排，不直接访问表
  application/     # 用例服务，可与 logic 合并但职责要清楚
  domain/          # 实体、值对象、领域规则、端口接口
  dal/
    repository/    # repository 实现
    tables/        # ORM/SQL 模型
  adapters/        # 外部服务适配器
  svc/             # 依赖组装和生命周期管理
```

#### 3.2 依赖方向

```text
server -> logic/application -> domain ports
                                  ^
                                  |
                       dal/adapters implement ports
```

- `logic` 依赖 domain/application 接口，不直接 import `dal`。
- `dal` 只实现 repository 接口，不承担业务编排。
- `svc` 负责创建数据库连接、缓存、消息生产者、repository、domain service 和 RPC client，并注入给 logic。
- 中间件通过接口注入或在 transport 层注册，不把具体中间件对象塞进所有服务的 `ServiceContext`。

#### 3.3 优先改造顺序

建议先改 `user-rpc`，以此作为其他服务模板：

1. 定义 user domain service 和 repository port。
2. 将现有 `dal` 改为 repository 实现。
3. 在 `svc` 中组装 repository、token、MFA、OSS 等适配器。
4. 将 RPC logic 改为只调用 application/domain service。
5. 用同样模式迁移 video、interaction、communication 和 chat。

### 阶段 4：领域数据所有权和跨服务通信（3～6 周）

本阶段开始真正形成微服务边界。

建议的数据所有权：

| 服务 | 第一阶段负责的数据 |
| --- | --- |
| user | 用户基础信息、MFA、第三方登录绑定 |
| video | 视频基础信息、热门和统计汇总 |
| interaction | 点赞、评论、评论点赞 |
| communication/relation | 关注、粉丝、好友关系 |
| chat | 房间、成员、消息、未读和在线状态 |

实施要求：

- 第一阶段可继续共用 MySQL 实例，但服务代码只能访问自己所有的表。
- 他域数据通过 RPC 获取，批量场景必须提供 `BatchGet` 接口，避免 N+1 调用。
- 点赞数、评论数、播放量、未读数等允许最终一致的数据通过 Kafka 事件更新。
- 事件至少包含 `event_id`、`event_type`、`occurred_at`、`trace_id`、`user_id`、`business_id` 和 `payload`。
- 消费者必须幂等；重复消费不能重复增加计数或重复写入。
- 等边界稳定后，再将共享数据库拆为独立 database，最后再决定是否拆为独立 MySQL 实例。

### 阶段 5：独立部署和服务治理（2～4 周）

- 每个 API、RPC、WS、MQ consumer 使用独立 Dockerfile 和配置文件。
- `compose.yml` 按服务提供健康检查、资源限制、启动依赖和独立日志。
- 生产环境引入 etcd 服务发现、超时、重试、熔断和限流策略。
- 使用 trace ID 和 OpenTelemetry 打通 gateway -> RPC -> MQ 链路。
- 建立每个服务的 `/healthz`、关键业务指标、错误率、RPC 延迟和 Kafka lag 监控。
- CI 按目录变更构建受影响服务，支持独立镜像发布和回滚。

## 7. 网关迁移的建议配置

`etc/gateway.yaml` 建议只包含以下类别：

```yaml
Name: gateway-api
Host: 0.0.0.0
Port: 8888
MaxBytes: 10485760

Auth:
  AccessSecret: ${ACCESS_SECRET}
  AccessExpire: 3600

UserRpc:
  Etcd:
    Hosts: [etcd:2379]
    Key: user.rpc

VideoRpc:
  Etcd:
    Hosts: [etcd:2379]
    Key: video.rpc

InteractionRpc:
  Etcd:
    Hosts: [etcd:2379]
    Key: interaction.rpc

CommunicationRpc:
  Etcd:
    Hosts: [etcd:2379]
    Key: communication.rpc

ChatRpc:
  Etcd:
    Hosts: [etcd:2379]
    Key: chat.rpc
```

数据库 DSN、Kafka brokers、OSS 密钥和 AI key 应从网关配置中移除，分别下沉到实际使用它们的服务。

## 8. 版本、测试和发布策略

### 8.1 分支和提交

- 每个阶段使用独立迁移分支，例如 `refactor/gateway-externalize`、`refactor/pkg-common`。
- 目录移动、依赖修改和行为修改分开提交，便于审查和回滚。
- 每次提交必须通过 `go test ./...`、`go vet ./...` 和 lint 检查。

### 8.2 测试分层

- 网关：路由、参数校验、鉴权、错误映射、RPC mock 和响应兼容测试。
- RPC：domain/application 单元测试、repository 集成测试、proto 契约测试。
- MQ：事件 schema、重复消费、失败重试和死信测试。
- 端到端：登录、发布视频、点赞、评论、关注、聊天和 WebSocket 主链路。
- 迁移期间对旧单体和新网关执行同一套 HTTP 回归用例，比较状态码、响应字段和业务结果。

### 8.3 发布切换

采用灰度和开关切换：

```text
旧 tiktok-api
      \            （按环境/用户/流量比例切换）
       -> gateway-api -> RPC
```

切换期间保留旧服务和旧镜像，发现错误时将入口流量切回旧服务。确认新网关稳定后，再下线旧 `tiktok.go` 和根级 API 启动配置。

## 9. 风险和应对措施

| 风险 | 影响 | 应对措施 |
| --- | --- | --- |
| 根级 `internal` 被多个服务引用 | 移动后编译大面积失败 | 先建立 `pkg` 兼容层，分批修改 import，禁止一次性删除 |
| 网关误保留数据库访问 | 服务边界继续恶化 | 在 CI 中检查 gateway 包禁止引用 GORM、SQL、DAL 和基础设施实现 |
| 跨服务共享表 | 数据一致性和发布互相阻塞 | 先定义表所有权，再用 RPC/MQ 替代直接访问 |
| RPC 超时或服务不可用 | API 延迟和错误增加 | 设置 deadline、有限重试、熔断、降级响应和监控告警 |
| API 响应不兼容 | 客户端故障 | 建立旧/新网关回归对比测试，保持路径和字段兼容 |
| WebSocket 迁移复杂 | 长连接中断 | 最后迁移 WS，先保证 HTTP 网关和 chat-rpc 稳定 |
| 配置和密钥散落 | 部署失败或泄露 | 网关、RPC、MQ consumer 分别维护配置，敏感值使用环境变量或密钥管理系统 |

## 10. 里程碑和验收标准

### M1：网关可独立运行

- `app/gateway/api` 独立启动并监听 8888。
- 根级 API 不再作为默认生产入口。
- 网关只依赖 RPC 和公共契约，HTTP 回归通过。

### M2：公共包完成迁移

- `pkg/ctxkey`、`pkg/xerr`、`pkg/jwt`、`pkg/id`、`pkg/mq` 和 `pkg/utils` 可独立测试。
- 新代码不再引用根级公共 `internal`。
- 根级 `internal` 仅剩待迁移的服务私有兼容代码，或已删除。

### M3：服务内聚完成

- 每个 RPC 的 logic 不直接依赖 DAL。
- 每个服务的 `svc` 只组装本服务依赖。
- RPC 接口、repository 接口和领域服务职责清晰。

### M4：数据边界完成

- 每张业务表有唯一服务所有者。
- 跨服务数据访问全部通过 RPC 或事件。
- 统计类数据具备幂等消费和重试机制。

### M5：独立部署完成

- 网关和各 RPC 服务可以独立构建、发布、扩容和回滚。
- 具备健康检查、日志、指标、trace ID 和基础告警。

## 11. 推荐执行顺序

```text
基线测试与依赖盘点
  -> 新建 app/gateway/api
  -> 网关接入 user-rpc
  -> 网关接入 video-rpc
  -> 网关接入 interaction/communication/chat-rpc
  -> gateway 灰度替换根级 tiktok.go
  -> internal 公共代码迁移到 pkg
  -> 以 user-rpc 为模板整改各服务 svc/logic/dal
  -> 明确表所有权，切断跨服务直连数据库
  -> Kafka 事件化统计和异步流程
  -> 独立部署、观测和后续数据库拆分
```

## 12. 结论

本项目不适合直接把现有目录复制成多个完全独立仓库。最稳妥的路径是先在 monorepo 中完成“网关外置 + RPC 契约化 + 公共包收敛 + 服务内聚”，再逐步完成数据和部署边界。第一阶段的成功标准不是目录数量增加，而是：网关不再持有业务基础设施、RPC 服务拥有自己的依赖组装、公共代码可以被安全复用、跨领域访问只能通过明确契约完成。

## 13. 第一阶段实施结果

截至当前版本，第一阶段迁移已落地：

- 根级 `internal`、根级 `tiktok.go` 和根级 `Dockerfile` 已移除。
- HTTP 网关统一位于 `app/gateway/api`，使用现有 go-zero `.api` 脚手架生成的 handler、logic、types 和 routes。
- RPC 领域类型位于 `pkg/contract`，HTTP DTO 仅保留在网关私有 `internal/types`。
- JWT、错误、上下文键、消息契约和工具函数位于 `pkg`。
- OSS 适配器位于 `pkg/storage/aliyun`；MFA 适配器位于 `app/user/rpc/internal/mfa`。
- 聊天缓存和 Kafka 实现位于 `app/chat/rpc/internal/infra`，不再由根级公共目录承载。
- Compose 和 Makefile 已切换到独立 `gateway` 服务及各 RPC 服务。

后续新增代码不得重新创建根级 `internal`；服务私有实现应放在对应 `app/<domain>/<service>/internal`，稳定的无业务归属能力才进入 `pkg`。
