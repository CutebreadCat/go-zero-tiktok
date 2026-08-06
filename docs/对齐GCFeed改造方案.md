# go-zero-tiktok 向 GCFeed 看齐:三项目对比与改造方案

> 目标:让 go_zero-tiktok 在工程化完备度上对齐 GCFeed(模板先进工程),fzuhelper-server 作为微服务可观测性参考。
> 技术栈保持 go-zero 微服务不变,借鉴的是 GCFeed 的**分层规范 / 文档体系 / 监控 / 测试 / 规格驱动**。

---

## 一、三项目横向对比

| 维度 | GCFeed(模板) | go-zero-tiktok(现状) | fzuhelper-server(参考) |
|---|---|---|---|
| 架构形态 | Gin 单体,DDD 四层 | go-zero 微服务(网关+RPC×4) | Kitex 微服务(网关+RPC×10) |
| 分层 | Domain/App/Infra/Interfaces 目录化 | rpc 内 DDD 化(config/dal/domain/logic/svc) | 三层(handler/service/pack)+pkg 下沉 |
| 文档体系 | **完整**(product/architecture/engineering/quickread/optimization/modules) | 残缺/过时(api.md/directory.md/mq.md) | 完整(development-guide/README.zh) |
| 工程规范文档 | **engineering.md 15节**(分层/命名/API/DB/错误/测试) | 无 | 有(开发指南+CODEOWNERS) |
| 监控 | **Prometheus 12指标 + Grafana 面板** | 无埋点(仅自定义 Zap 日志) | OTel trace/metric/log 三合一 |
| 链路追踪 | 无(单体) | 无(go-zero 有传递依赖未启用) | **jaeger/OTel 完整** |
| 测试 | 模块 API 测试(内存仓储mock)+worker测试 | **零测试** | 数据层+middleware 大量单测 |
| 数据库 | 表结构清晰,高频计数独立统计表 | migrations与GORM模型**不同步** | 分层清晰,数据层下沉 pkg |
| MQ | RabbitMQ worker(fanout/互动落库/embedding) | Kafka 库**已引入未接线** | Kafka 完整(日志投递等) |
| 规格驱动 | **OpenSpec**(change→spec→validate) | 无 | 无 |
| CI/CD | 无 CI(本地验证) | 仅 golangci-lint | 完整 CI+CD |
| 配置 | YAML | go-zero conf + viper(aliyun) | etcd3 远程 + k8s ConfigMap |

**结论**:GCFeed 强在**工程规范文档化 + 监控 + 测试 + 规格驱动**;fzuhelper-server 强在**微服务可观测性(OTel)**。go-zero-tiktok 目前是"壳子齐全但工程化空洞"——分层骨架在,但缺文档、缺测试、缺监控、缺接线。

---

## 二、接入优先级(按 ROI 排序)

### P0 — 地基(先做,不做则上层无从谈起)
1. **数据库迁移与代码对齐** — 统一 `migrations/000001` 目标态 与 GORM 模型(video_popular→video_stat、user_mfa 合并、补 playback_qos_reports)。
2. **工程规范文档化** — 新建 `docs/engineering.md`(分层/命名/API风格/DB规范/错误码/测试约定),让新代码有据可依。
3. **建立单元/接口测试** — 至少覆盖核心模块(用户/视频/互动/关注)成功+参数错误+鉴权+幂等+游标路径。

### P1 — 工程完善
4. **Prometheus 指标埋点 + Grafana** — 网关 HTTP 计数器/直方图;视频/互动业务指标。
5. **链路追踪** — 启用 go-zero 的 OTel/jaeger(参考 fzuhelper-server 的 OTel 方案)。
6. **Kafka 接线** — 把 `pkg/kafka` 用起来:视频发布事件、互动异步落库(参考 GCFeed 的 worker 模式)。

### P2 — 进阶
7. **OpenSpec 规格驱动** — 引入 openspec 基线,新功能先建 change 再实现。
8. **Feed 性能优化** — 页缓存/singleflight/热榜分钟桶/关注流 inbox-outbox(GCFeed optimization.md 直接可迁移)。
9. **日志体系规范化 / 文档体系进一步铺全**(quickread/architecture/uiux/interview)。

---

## 三、各阶段改造细节

### 阶段 A:数据库对齐(P0-1)
- 现状:`migrations/000001` 是目标态,video 服务还在用 `video_popular`,`user_mfa` 独立表。
- 动作:改 GORM 模型对齐迁移;`dal/tables/video_popular` → `video_stat`;`user_baseinfo` 合并 MFA 字段;新增 `playback_qos_reports` 模型。
- 参考:GCFeed 高频计数独立成统计表(`video_stat`/`user_relation_stat`),go-zero-tiktok 迁移已采用该模式,补齐代码即可。

### 阶段 B:工程规范文档(docs/engineering.md)(P0-2)
内容对齐 GCFeed engineering.md 但适配 go-zero:
- 服务与目录规则(api/ + app/{svc}/rpc/internal/{config,dal,domain,logic,svc})
- 命名规范(包名/repo/domain)
- REST API 风格(方法语义/路径/状态码/游标分页/幂等键)
- 数据库规范(小写蛇形/主表单数/关系表业务名/通用字段)
- 错误处理(统一 xerr)
- 测试约定(必测路径)

### 阶段 C:测试体系(P0-3)
- 参考 GCFeed 的"内存仓储 mock 驱动"思路,但用 go-zero 的 gock/mockery 或手写接口桩。
- 每个 RPC 服务:`$svc/rpc` 的 logic 补单测;网关 handler 补接口测试。
- Makefile `test` 目标现在空跑,修成真正跑 `go test ./...`。

### 阶段 D:监控(P1-4)
- 新增 `pkg/metrics`(参考 GCFeed metrics.go 的 12 指标,适配 go-zero)。
- 网关中间件埋 `http_requests_total` / `http_request_duration_seconds`。
- 视频上传/处理、互动、worker 业务指标。
- 补 Prometheus 抓取 target + Grafana 面板(json)。

### 阶段 E:链路追踪(P1-5)
- 参考 fzuhelper-server:`tracing.NewOtelProvider(serviceName, endpoint)` + 注册 shutdown hook。
- go-zero 侧:启用 `otel` 中间件/拦截器,配置 jaeger/OTel collector 地址。

### 阶段 F:MQ 接线(P1-6)
- `pkg/kafka` 已实现 Producer/Consumer/多主题单元/工作池,缺接线。
- 参考 GCFeed worker:视频发布→发事件;互动→异步落库;可加 embedding/推荐任务。
- 补 worker 进程或并入现有服务。

### 阶段 G:OpenSpec(P2-7)
- 复制 GCFeed openspec 结构:`project.md` + `changes/`。
- 新功能先 `openspec new <change>` → proposal/design/tasks/spec → validate。

### 阶段 H:Feed 优化(P2-8)
- 迁移 GCFeed optimization.md 的 P0-01~P0-07:页缓存、singleflight、批量 MGET、热榜分钟桶、inbox-outbox fanout。

---

## 四、建议执行顺序(里程碑)

```
M1 数据库对齐 + 工程规范文档 + 核心测试(地基)
M2 监控(Prometheus+Grafana)+ 链路追踪(OTel)
M3 Kafka 接线 + MQ worker
M4 OpenSpec + Feed 优化
```

每个里程碑独立可交付、可演示。M1 是其它一切的前置。

---

## 五、文件改动清单(仅 go-zero-tiktok)

新增:
- `docs/engineering.md`(工程规范)
- `docs/architecture.md`、`docs/quickread.md`(补全文档体系)
- `pkg/metrics/`(指标定义+中间件)
- `docs/modules/`(模块设计文档)
- `openspec/`(规格驱动基线)
- `ops/grafana/`(面板 json+Pometheus 配置)

修改:
- 各 `app/{svc}/rpc/internal/dal/tables/*/model.go`(数据库对齐)
- `app/video/rpc`(video_popular→video_stat)
- `app/user/rpc`(user_mfa 合并)
- `app/gateway/api/internal/handler/`(指标中间件)
- `Makefile`(test 目标修复)
- 各服务 main.go(OTel provider 接入)