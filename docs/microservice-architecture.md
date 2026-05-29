# go-zero 微服务拆分分析

本文基于当前 `go_zero-tiktok` 项目的代码结构，分析如何从现有单体架构拆解为 go-zero 微服务体系，并对比单体架构与微服务架构的优缺点。

## 1. 当前项目架构现状

当前项目更准确地说是一个“模块化单体”。代码中已经按照业务能力拆出了 `user`、`video`、`interaction`、`communication`、`chat`、`websocket` 等领域目录，但这些模块仍然运行在同一个进程中，共享同一个 `ServiceContext`、同一个主程序入口、同一套配置和同一套数据库连接。

当前主链路大致如下：

```text
客户端
  -> tiktok-api HTTP 服务
    -> handler
      -> logic
        -> domain service
          -> repository
            -> MySQL / Redis / Kafka / OSS / AI 等基础设施
```

这套结构的优点是开发简单、调用直接、调试方便。缺点是所有业务被打包成一个服务发布，某个模块的故障、性能瓶颈、依赖变更，都可能影响整个系统。

当前比较明显的领域边界如下：

| 当前模块 | 主要能力 | 相关目录 |
| --- | --- | --- |
| 用户模块 | 注册、登录、刷新 token、用户信息、头像、MFA、教务系统登录 | `api/user`、`internal/domain/user`、`internal/logic/user` |
| 视频模块 | 投稿、视频列表、搜索、推荐 feed、热门视频、视频基础信息 | `api/video`、`internal/domain/video`、`internal/logic/video` |
| 互动模块 | 视频点赞、评论发布、评论列表、评论点赞、删除评论 | `api/interaction`、`internal/domain/comment`、部分 `internal/domain/video` |
| 关系模块 | 关注、粉丝、关注列表、好友列表 | `api/communication`、`internal/domain/userfollow` |
| 聊天模块 | 聊天房间、消息列表、加入房间、WebSocket 聊天 | `api/chat`、`internal/domain/chat`、`internal/domain/websocket` |
| 基础设施 | MySQL、Redis、Kafka、阿里云 OSS、AI Agent、限流、熔断 | `internal/infra`、`internal/middleware`、`internal/dal` |

## 2. 单体架构与微服务架构的核心区别

单体架构不是“代码不分层”，而是“部署单元是一个”。你现在的项目虽然代码分了模块，但最终仍然是一个 `tiktok-api` 进程，所以它属于单体。

微服务架构的核心不是简单把目录拆开，而是把业务能力拆成多个可以独立部署、独立扩缩容、独立演进的服务。每个服务拥有自己的入口、配置、依赖注入、数据库访问层和服务契约。服务之间不再直接调用 Go 代码，而是通过 RPC、HTTP 或消息队列通信。

对比关系如下：

```text
单体架构：

客户端
  -> 一个 tiktok-api
    -> user/video/comment/chat/follow 等模块
      -> 共享数据库、Redis、Kafka

微服务架构：

客户端
  -> API 网关 / BFF
    -> user-api / video-api / interaction-api / chat-api
      -> user-rpc / video-rpc / comment-rpc / relation-rpc / chat-rpc
        -> 各服务自己的数据表、缓存、消息队列
```

在 go-zero 体系中，常见做法是：

- `api` 服务负责对外 HTTP 接口、鉴权、参数校验、聚合返回。
- `rpc` 服务负责内部业务能力暴露，供其他服务调用。
- `mq` / `job` 服务负责异步任务、事件消费、定时任务。
- `model` / `repository` 层尽量归属到对应服务内部，避免跨服务直接读写别人的表。

## 3. 建议的服务拆分方案

结合当前业务规模，不建议一开始拆得太碎。推荐按“业务边界 + 数据所有权 + 调用频率”拆成 5 到 7 个服务。

### 3.1 用户服务 user

职责：

- 用户注册、登录、刷新 token。
- 用户基础信息查询。
- 用户头像上传。
- MFA 绑定与校验。
- 教务系统账号绑定或登录能力。

建议拆分：

```text
app/user/api
app/user/rpc
```

对外 HTTP：

- `/user/register`
- `/user/login`
- `/user/token/refresh`
- `/user/info`
- `/user/avatar/upload`
- `/user/mfa/qrcode`
- `/user/mfa/bind`
- `/user/jwch/login`
- `/user/jwch/cookie`

内部 RPC 能力：

- `GetUser(ctx, userId)`
- `BatchGetUsers(ctx, userIds)`
- `ValidateUser(ctx, userId)`
- `UpdateAvatar(ctx, userId, avatarUrl)`

数据归属：

- `user_baseinfo`
- 用户 MFA 字段
- 用户第三方绑定信息

其他服务如果需要用户信息，不应该直接查 `user_baseinfo` 表，而应该调用 `user-rpc`。

### 3.2 视频服务 video

职责：

- 视频发布。
- 视频基础信息查询。
- 作者视频列表。
- feed 流。
- 视频搜索。
- 热门视频。
- 视频播放、点赞、评论等统计信息维护。

建议拆分：

```text
app/video/api
app/video/rpc
app/video/mq
```

对外 HTTP：

- `/video/publish`
- `/video/list`
- `/video/search`
- `/video/popular`
- `/video/feed`

内部 RPC 能力：

- `PublishVideo(ctx, req)`
- `GetVideo(ctx, videoId)`
- `BatchGetVideos(ctx, videoIds)`
- `ListVideosByAuthor(ctx, authorId, page)`
- `IncreaseVisitCount(ctx, videoId, delta)`
- `UpdateLikeCount(ctx, videoId, delta)`
- `UpdateCommentCount(ctx, videoId, delta)`

数据归属：

- `video_baseinfo`
- `video_popular`

建议把当前 `PopularRepo` 放入 `video-rpc` 所属的数据访问层。评论服务、点赞服务不要直接操作 `video_popular`，而是通过 `video-rpc` 或事件通知视频服务更新计数。

### 3.3 互动服务 interaction

当前 `interaction` 同时包含点赞和评论。初期可以先作为一个互动服务，后续量大时再拆成 `like` 和 `comment` 两个服务。

职责：

- 视频点赞、取消点赞。
- 用户点赞列表。
- 评论发布。
- 评论回复。
- 评论列表。
- 评论删除。
- 评论点赞。

建议拆分：

```text
app/interaction/api
app/interaction/rpc
app/interaction/mq
```

对外 HTTP：

- `/like/action`
- `/like/list`
- `/comment/publish`
- `/comment/list`
- `/comment/delete`
- `/comment/like`
- `/comment/parent`

内部 RPC 能力：

- `LikeVideo(ctx, userId, videoId, actionType)`
- `GetLikedVideoIds(ctx, userId, page)`
- `CreateComment(ctx, userId, videoId, content)`
- `ReplyComment(ctx, userId, parentCommentId, content)`
- `DeleteComment(ctx, userId, commentId)`
- `ListComments(ctx, videoId, page)`

数据归属：

- `video_liker`
- `comment_baseinfo`
- `comment_liker`

跨服务关系：

- 点赞视频成功后，通知 `video-rpc` 更新点赞数，或发送 `video_liked` 事件由 `video-mq` 异步更新。
- 评论成功后，通知 `video-rpc` 更新评论数，或发送 `comment_created` 事件由 `video-mq` 异步更新。
- 获取点赞列表时，互动服务返回 videoId 列表，API 层或 interaction-rpc 再调用 `video-rpc` 批量获取视频详情。

建议优先用事件异步更新统计值，因为点赞数、评论数这类数据允许短暂最终一致。

### 3.4 关系服务 relation

当前目录名是 `communication`，领域名是 `userfollow`。微服务中建议命名为 `relation`，更贴近关注、粉丝、好友关系。

职责：

- 关注、取消关注。
- 关注列表。
- 粉丝列表。
- 好友列表。

建议拆分：

```text
app/relation/api
app/relation/rpc
```

对外 HTTP：

- `/relation/action`
- `/following/list`
- `/follower/list`
- `/friend/list`

内部 RPC 能力：

- `Follow(ctx, userId, toUserId)`
- `Unfollow(ctx, userId, toUserId)`
- `GetFollowingIds(ctx, userId, page)`
- `GetFollowerIds(ctx, userId, page)`
- `GetFriendIds(ctx, userId, page)`
- `IsFollowing(ctx, userId, targetUserId)`

数据归属：

- `user_follow`

跨服务关系：

- 关系列表展示需要用户信息时，通过 `user-rpc.BatchGetUsers` 批量补全用户资料。
- 聊天服务创建私聊房间时，可以调用 `relation-rpc` 判断是否为好友。

### 3.5 聊天服务 chat

职责：

- 聊天房间创建。
- 房间列表。
- 加入房间。
- 历史消息查询。
- WebSocket 连接管理。
- 消息投递、未读数、在线状态。

建议拆分：

```text
app/chat/api
app/chat/rpc
app/chat/ws
app/chat/mq
```

对外 HTTP：

- `/chat/room/create`
- `/chat/rooms`
- `/chat/messages`
- `/chat/room/join`

WebSocket：

- `/chat/ws`

内部 RPC 能力：

- `CreateRoom(ctx, req)`
- `JoinRoom(ctx, req)`
- `GetRooms(ctx, userId)`
- `GetMessages(ctx, roomId, page)`
- `SaveMessage(ctx, message)`

数据归属：

- chat room 表
- message 表
- room member 表

缓存归属：

- WebSocket 在线状态
- 房间连接关系
- 未读消息

跨服务关系：

- 创建私聊或群聊时，可以调用 `user-rpc` 校验用户存在。
- 如果要求只有好友能聊天，则调用 `relation-rpc` 校验关系。
- WebSocket 收到消息后，可以写入 Kafka，由 `chat-mq` 持久化和分发。

### 3.6 AI 服务 ai

当前 AI 能力在 `internal/infra/ai`，并被 websocket/chat 使用。建议作为后期拆分项，不必第一阶段就拆。

职责：

- AI Agent 调用。
- AI 聊天消息处理。
- 限流、熔断、外部模型适配。

建议拆分：

```text
app/ai/rpc
app/ai/mq
```

内部 RPC 能力：

- `Chat(ctx, messages)`
- `GenerateReply(ctx, req)`

跨服务关系：

- `chat` 服务将 AI 消息请求发给 `ai-rpc`。
- 如果 AI 响应慢，推荐通过 MQ 异步处理，避免阻塞 WebSocket 主链路。

## 4. 推荐的 go-zero 目录结构

拆成微服务后，可以采用如下结构：

```text
go_zero-tiktok/
  app/
    user/
      api/
      rpc/
      model/
    video/
      api/
      rpc/
      mq/
      model/
    interaction/
      api/
      rpc/
      mq/
      model/
    relation/
      api/
      rpc/
      model/
    chat/
      api/
      rpc/
      ws/
      mq/
      model/
    ai/
      rpc/
      mq/
  common/
    ctxkey/
    xerr/
    jwt/
    middleware/
    mq/
    id/
  deploy/
    docker-compose/
    k8s/
  docs/
```

也可以先保留当前 `internal/domain` 的领域代码，把服务入口逐步迁出去。不要一上来把所有代码复制成多个仓库。第一阶段更适合“单仓多服务”，也就是 monorepo。

## 5. go-zero 微服务中的调用方式

go-zero 中常见链路如下：

```text
外部请求
  -> xxx-api
    -> logic
      -> xxx-rpc client
        -> xxx-rpc
          -> domain service
            -> repository/model
```

举例：用户获取好友列表。

```text
客户端
  -> relation-api /friend/list
    -> relation-rpc.GetFriendIds(userId)
    -> user-rpc.BatchGetUsers(friendIds)
    -> 返回好友用户信息
```

举例：发布评论。

```text
客户端
  -> interaction-api /comment/publish
    -> interaction-rpc.CreateComment(userId, videoId, content)
      -> 写 comment_baseinfo
      -> 发送 comment_created 事件
    -> video-mq 消费 comment_created
      -> 更新 video_popular.comment_count
```

同步 RPC 适合强依赖、需要立即返回的数据，例如获取用户资料、获取视频详情。MQ 适合最终一致的数据，例如播放量、点赞数、评论数、消息推送、未读数。

## 6. 数据库如何拆

微服务拆分中最容易出问题的是数据库边界。

当前项目所有 repository 都在一个进程内，理论上可以任意查任意表。但微服务中应遵守一个原则：

每个服务只直接读写自己的表，其他服务的数据通过 RPC 或事件获得。

建议数据归属如下：

| 服务 | 拥有的数据 |
| --- | --- |
| user | `user_baseinfo`、MFA、第三方登录绑定 |
| video | `video_baseinfo`、`video_popular` |
| interaction | `video_liker`、`comment_baseinfo`、`comment_liker` |
| relation | `user_follow` |
| chat | chat room、message、member、unread、presence |

第一阶段可以物理上仍然共用一个 MySQL 实例，甚至共用一个 database，但代码层面要先做到“服务不跨库表访问”。等边界稳定后，再拆成多个 database 或多个 MySQL 实例。

## 7. 服务间一致性设计

你这个项目里最适合用最终一致的场景有：

- 视频播放量增加。
- 视频点赞数增加或减少。
- 视频评论数增加或减少。
- 聊天未读数更新。
- 热门视频榜单更新。

这些数据不需要在用户请求返回时绝对准确，允许几百毫秒到几秒延迟。因此推荐通过 Kafka 事件处理。

建议事件：

```text
video_viewed
video_liked
video_unliked
comment_created
comment_deleted
message_sent
room_joined
```

事件消息中至少包含：

```text
event_id
event_type
occurred_at
trace_id
user_id
business_id
payload
```

消费者要注意幂等。例如 `event_id` 已处理过就跳过，避免 Kafka 重试导致计数重复增加。

## 8. 推荐迁移步骤

### 第一阶段：整理边界，不拆进程

目标是让当前单体更像“可拆分单体”。

- 保持当前一个 `tiktok-api` 进程。
- 为每个领域补充接口端口，避免跨领域直接依赖具体实现。
- `comment`、`relation`、`chat` 等模块不要直接使用其他领域 repository。
- 明确每张表归属哪个领域。
- 把公共能力沉淀到 `common` 或 `shared`，例如错误码、上下文 key、JWT、ID 生成。

你刚刚对 `comment` 做的接口拆分，就是这个阶段的正确方向。

### 第二阶段：先拆 user-rpc 和 video-rpc

优先拆 `user` 和 `video`，因为它们是其他服务最常依赖的基础能力。

- 用 `.proto` 定义 `user-rpc`。
- 用 `.proto` 定义 `video-rpc`。
- 当前 API 层逐步改为调用 RPC client。
- 保留原来的 HTTP 路由不变，对前端无感。

### 第三阶段：拆 interaction-rpc

把点赞和评论相关数据访问收敛到 interaction 服务。

- `video_liker`、`comment_baseinfo`、`comment_liker` 归 interaction。
- 点赞、评论成功后，通过 RPC 或 MQ 通知 video 更新统计。
- 点赞列表需要视频详情时调用 `video-rpc.BatchGetVideos`。

### 第四阶段：拆 relation-rpc

关注关系相对独立，适合单独拆。

- 关注、粉丝、好友列表只由 relation 服务查表。
- 返回用户详情时调用 `user-rpc.BatchGetUsers`。

### 第五阶段：拆 chat/ws

聊天和 WebSocket 的运行模型与普通 HTTP API 不同，适合后拆。

- 先拆普通 chat-api/chat-rpc。
- 再拆 WebSocket 网关。
- 消息写入、离线推送、未读数通过 MQ 消费。

### 第六阶段：服务治理

当多个服务真正独立运行后，需要补齐治理能力：

- 服务注册与发现：etcd。
- 链路追踪：trace id、OpenTelemetry。
- 日志聚合：按 trace id 查询一次请求经过的所有服务。
- 熔断限流：go-zero 内置中间件 + 自定义限流。
- 配置管理：按服务拆分配置。
- CI/CD：每个服务可以独立构建镜像和发布。

## 9. 单体架构优缺点

优点：

- 开发简单，一个进程、一个配置、一个启动入口。
- 本地调试方便，不需要同时启动很多服务。
- 函数调用成本低，不需要处理网络失败、超时、序列化。
- 事务处理简单，跨模块操作可以直接使用同一个数据库事务。
- 早期开发效率高，适合需求变化快、团队规模小的阶段。

缺点：

- 所有模块一起发布，一个小改动也要重新部署整个服务。
- 模块间边界容易变模糊，repository 和 domain 容易互相调用。
- 某个模块流量暴涨时，只能整体扩容，资源利用率低。
- 某个模块故障可能拖垮整个进程。
- 代码规模变大后，编译、测试、发布、协作成本都会上升。

## 10. 微服务架构优缺点

优点：

- 服务可以独立发布，用户、视频、互动、聊天可以分开迭代。
- 可以按热点模块单独扩容，例如只扩容 chat/ws 或 video/feed。
- 故障隔离更好，评论服务故障不应该影响登录服务。
- 领域边界更清晰，每个服务拥有自己的数据和业务规则。
- 更适合多人团队并行开发。
- 可以针对不同服务选择不同资源配置，例如 WebSocket 服务更关注连接数，视频服务更关注存储和搜索。

缺点：

- 系统复杂度明显上升，需要服务发现、配置、日志、链路追踪、监控告警。
- 服务间调用会出现网络超时、重试、熔断、降级等问题。
- 本地开发成本更高，需要启动多个 api/rpc/mq 服务。
- 数据一致性更复杂，很多场景不能再依赖单数据库事务。
- 接口契约维护成本更高，`.proto`、`.api`、事件消息都需要版本管理。
- 排查问题更难，一次请求可能经过多个服务。

## 11. 对当前项目的推荐结论

当前项目不建议立刻完全微服务化。更好的路线是：

```text
模块化单体
  -> 边界清晰的可拆分单体
    -> user/video 基础服务 RPC 化
      -> interaction/relation/chat 逐步独立
        -> MQ 事件化统计与消息链路
```

也就是说，先不要急着拆部署单元，而是先把依赖方向整理干净：

- API 层只做协议适配和参数校验。
- logic 层编排业务流程。
- domain service 持有业务规则。
- repository 只归属本服务自己的数据。
- 跨领域调用通过接口端口表达。
- 跨服务调用将来替换为 go-zero RPC client。
- 统计、通知、消息这类副作用优先事件化。

如果按这个路线走，你的项目既能保留当前单体开发的效率，又能为后续 go-zero 微服务拆分留下清晰的落点。

