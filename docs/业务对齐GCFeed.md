# 业务对齐 GCFeed 方案

> 目标:让 go-zero-tiktok 在**业务能力**上对齐 GCFeed(短视频 Feed 系统参考工程)。
> 技术栈保持 go-zero 微服务不变,借鉴的是业务模块划分与能力清单。

---

## 一、业务模块对照(GCFeed 12 模块 vs go-zero-tiktok)

| GCFeed 模块 | 状态 | go-zero-tiktok 现状 | 差距 |
|---|---|---|---|
| 账户 account | ✅ | 注册/登录/刷新/资料/头像/MFA | 基本齐,缺登出 |
| 视频 video | ✅ | 发布/feed/list/search/popular/点赞 | 齐 |
| Feed feed | ✅ | GetFeedVideo/GetPopularVideo | 缺推荐流 scene、游标分页、打散 |
| 推荐 recommendation | ✅ | 无 | 整个缺失(召回/排序/曝光去重) |
| 互动 interaction | ✅ | 评论/回复/评论点赞 | 缺收藏(favorite) |
| 关系 relation | ✅ | 关注/粉丝/关注列表/好友 | 齐 |
| 消息 message | ✅ | 无 | 整个缺失(通知/未读/已读) |
| 播放优化 playback | ✅ | 表已建(playback_qos_reports)无业务 | 表建了没接 |
| 审核 review | 🔲规划 | 无 | 两边都规划,不算差距 |
| 后台运营 admin | 🔲规划 | 无 | 同上 |
| 系统治理 governance | 🔲规划 | 无 | 同上 |
| 监控告警 monitoring | 🔲规划 | 无 | 同上 |

**结论**:go-zero-tiktok 用户端基础闭环已覆盖 GCFeed 8 个已实现模块中的 6 个,业务差距集中在三条链路。

## 二、真正的业务差距(按优先级)

### P0 — 短视频核心,不做不算短视频

**1. Feed 推荐流增强(video 服务)**
- 加 scene 分流(推荐 / 关注 / 最新)
- 引入游标分页(当前是 page_num/page_size 传统分页;目标 `{items, next_cursor, has_more}`)
- 曝光去重,避免刷到重复视频

**2. 收藏(interaction 服务)**
- 点赞已有,补收藏非常轻:一张 `video_favorite` 表 + 两个接口
- REST:`PUT /videos/:id/favorite` / `DELETE /videos/:id/favorite`

### P1 — 体验闭环

**3. 消息中心(新业务域)**
- 站内通知:关注/点赞/评论/收藏触发通知、未读数、批量已读
- GCFeed 用"事件生成消息"模式,正好接 `pkg/kafka`(已引入未接线)
- 参考:message/event 驱动,消费互动事件写消息

**4. 播放质量上报落地(video 服务)**
- `playback_qos_reports` 表和模型已建好,接上报接口即可
- REST:`POST /playback-qos-reports`

### P2 — 内容治理(GCFeed 也还在规划,可后置)

**5. 审核 / 后台运营 / 系统治理** — 两边都规划中,不急着追

## 三、接口映射参考(GCFeed 资源化风格)

| 能力 | GCFeed 路径 | go-zero-tiktok 落地 |
|---|---|---|
| 收藏 | PUT/DELETE `/api/videos/{id}/favorite` | `interaction` 组,新表 video_favorite |
| 消息列表 | GET `/api/messages` | 新服务或 communication |
| 未读计数 | GET `/api/message-stats/unread` | 同上 |
| 批量已读 | PATCH `/api/messages` | 同上 |
| 播放上报 | POST `/api/playback-qos-reports` | video 组,复用已建模型 |
| 推荐候选 | POST `/internal/recommendation-candidates` | 网关内部接口,走推荐服务 |

## 四、执行顺序

```
M1 收藏 + 播放上报落地(改动小、风险低,先做)
M2 Feed 推荐流 + 游标分页(video 服务增强)
M3 消息中心(新业务域 + Kafka 事件)
M4 审核/后台运营(GCFeed 对齐后再追)
```

每步独立可交付。M1 是零架构改动、价值最直接。

---

## 附:GCFeed 业务模块文档地图

GCFeed 每模块有统一文档模板(职责/接口/数据表/业务规则/测试/前端接入点),见 `docs/modules/*.md`,新增业务模块时可参考其规格写法。