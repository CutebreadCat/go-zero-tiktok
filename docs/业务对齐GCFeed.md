# 业务对齐 GCFeed 方案

> 目标:让 go-zero-tiktok 在**业务能力**上对齐 GCFeed(短视频 Feed 系统参考工程)。
> 技术栈保持 go-zero 微服务不变,借鉴的是业务模块划分与能力清单。

| 元信息 | 内容 |
|---|---|
| 状态 | **已修订**（2026-08-28 按代码实际进度更新） |
| 上次更新 | 2026-08-28 |
| 适用范围 | 业务能力对照基线 |
| 下一步规划 | 见 [下一阶段发展规划-推荐飞轮与QoE.md](./下一阶段发展规划-推荐飞轮与QoE.md) |

---

## 一、业务模块对照(GCFeed 12 模块 vs go-zero-tiktok)

| GCFeed 模块 | 状态 | go-zero-tiktok 现状 | 差距 |
|---|---|---|---|
| 账户 account | ✅ | 注册/登录/刷新/资料/头像/MFA | 基本齐,缺登出 |
| 视频 video | ✅ | 发布/作者列表/搜索/热门 | 缺视频详情单条 API、编辑/删除/下架、转码与封面 |
| Feed feed | ✅ | timeline/following/hot/recommend 四种 scene,游标分页,曝光去重,同作者打散 | 缺复合游标 tie-breaker、inbox 容量治理 |
| 推荐 recommendation | 🟡 部分 | 规则版推荐已落地（三路召回+规则粗排+QoS 加权+曝光去重） | 缺画像/负反馈/解释字段/实验；见 Phase 2 文档 |
| 互动 interaction | ✅ | 点赞/收藏/评论/回复/评论点赞 | 缺回复计数、评论点赞数维护、Redis 冷启动恢复 |
| 关系 relation | ✅ | 关注/粉丝/关注列表/好友 | `user_relation_stat` 表已建但未接入业务 |
| 消息 message | ✅ | 站内通知 LIKE/COMMENT/FOLLOW/SYSTEM、未读数、批量已读（bdeb051） | 缺收藏/回复类型、稳定 event_id、idempotency_key 未接线 |
| 播放优化 playback | 🟡 部分 | QoS 上报/聚合/QoS 参与排序已接入 | 聚合仅视频级维度；缺网络/设备分桶 |
| 审核 review | 🔲规划 | 无（仅 `report_count` 字段预留） | 两边都规划,不算差距 |
| 后台运营 admin | 🔲规划 | 无 | 同上 |
| 系统治理 governance | 🔲规划 | 无 | 同上 |
| 监控告警 monitoring | 🟡 部分 | 日志链路(Loki/Alloy/Grafana)已接;Prometheus/Trace 后端未部署 | 指标存储与可视化缺失,见监控方案文档 |

**结论(2026-08 修订)**:用户端基础闭环已全部覆盖 GCFeed 已实现模块；当前差距不再是"缺模块"，而是三类**深度缺口**——推荐反馈闭环、数据口径正确性、可观测性落地。详见 [下一阶段发展规划-推荐飞轮与QoE.md](./下一阶段发展规划-推荐飞轮与QoE.md)。

## 二、真正的业务差距(按优先级,2026-08 修订)

### P0 — 数据口径正确性(差异化改造的前置)

**1. served / impression 口径拆分(video 服务)**
- 当前推荐返回前即 `MarkSeen`(recommend_strategy.go),真实客户端 impression 消费时又写 seen,"下发"与"真实曝光"混用；
- 目标四层事件:`served → impression → play → complete`,CTR/画像/创作者统计只基于真实 impression。

**2. 事件稳定幂等 ID(全链路)**
- tracking 事件无去重;消息 event_id 用 `UnixMilli` 拼接,重试跨毫秒会重复通知；
- 目标:统一事件 envelope + `event_id` 数据库唯一约束,幂等键来自业务实体(comment_id/relation_id 等)。

**3. Feed 正确性修复(video 服务)**
- Redis 游标 `lastTimeMs+1` 开区间在同毫秒多视频时会漏项,需复合游标(score+video_id)；
- 热榜 `ZREMRANGEBYRANK` 裁剪方向疑似删除高分侧,需验证修正。

### P1 — 推荐反馈闭环(差异化主线)

**4. 画像/负反馈/可解释排序/轻量实验** — 详见 [下一阶段发展规划-推荐飞轮与QoE.md](./下一阶段发展规划-推荐飞轮与QoE.md) 与 [recommend-phase2-tracking-events-design.md](./recommend-phase2-tracking-events-design.md)。

**5. QoE 感知推荐(video 服务)**
- `video_qos_stat` 当前仅视频级聚合,需扩展网络/设备分桶 + 最小样本门槛 + 短期降权/自动恢复。

**6. 创作者内容健康诊断(video 服务)**
- 聚合 `video_view_events` + `video_qos_stat` + 互动统计,给作者暴露曝光/完播/卡顿/转化报告。

### P2 — 治理与工程化

**7. 审核 / 后台运营 / 系统治理** — 两边都规划中,不急着追;若做,建议绑定负反馈形成"举报→降权→审核→申诉"闭环,而不是孤立 CRUD。

## 三、接口映射参考(GCFeed 资源化风格)

| 能力 | GCFeed 路径 | go-zero-tiktok 落地 |
|---|---|---|
| 收藏 | PUT/DELETE `/api/videos/{id}/favorite` | `interaction` 组,复用 `video_interaction` 表 action_type=2 |
| 消息列表 | GET `/api/messages` | ✅ communication 域,`GET /messages` |
| 未读计数 | GET `/api/message-stats/unread` | ✅ `GET /message-stats/unread` |
| 批量已读 | PATCH `/api/messages` | ✅ `PATCH /messages` |
| 播放上报 | POST `/api/playback-qos-reports` | video 组,复用已建模型 |
| 推荐候选 | POST `/internal/recommendation-candidates` | 网关内部接口,走推荐服务 |

## 四、执行顺序(历史记录,已完成 M1-M4)

```
M1 ✅ 收藏已完成(含服务边界重构)
M2 ✅ 播放质量上报+聚合+QoS 参与排序已落地
M3 ✅ Feed 四 scene + 游标分页 + 规则推荐 + 曝光去重 + 同作者打散
M4 ✅ 消息中心基础版(站内通知/未读/批量已读)
M5 ⬜ 推荐反馈闭环(画像/负反馈/解释/实验) → 见下一阶段发展规划
M6 ⬜ QoE 感知 + 创作者诊断 → 见下一阶段发展规划
M7 ⬜ 审核/后台运营(建议绑定负反馈闭环后再做)
```

---

## 附:GCFeed 业务模块文档地图

GCFeed 每模块有统一文档模板(职责/接口/数据表/业务规则/测试/前端接入点),见 `docs/modules/*.md`。
> ⚠️ 2026-08 核对:`docs/modules/` 目录在本仓库**不存在**,上述模板仅作为 GCFeed 参考工程的说明;本仓库如需模块化文档,请按 docs/README.md 的状态规范自行建立。