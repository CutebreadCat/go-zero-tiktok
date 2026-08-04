---
name: database-reconstruction-plan
overview: 按照方案.md执行数据库重构：雪花ID改bigint、软删除方案落地（user_baseinfo用status，user_follow用active_flag生成列）、表归属收敛去重、幂等键添加、新增user_relation_stat表、引入golang-migrate迁移工具。
todos:
  - id: migrate-infra
    content: 搭建 golang-migrate 基础设施：安装依赖，创建 migrations 目录，编写初始基线迁移（从现有 schema.sql 拆解），封装 pkg/migrate 启动器
    status: completed
  - id: bigint-migration
    content: 雪花ID bigint 全量迁移：数据库层 ALTER 所有 ID 列为 BIGINT，GORM model 层 string→int64，proto 层 string→int64，契约层 int64+json,string，业务逻辑层移除字符串ID转换
    status: completed
    dependencies:
      - migrate-infra
  - id: user-status-softdelete
    content: user_baseinfo 软删除改造（status方案）：新增 status TINYINT 字段，唯一索引改为 (username, status)，所有查询加 status=1 过滤，删除改为 UPDATE status=0
    status: completed
    dependencies:
      - bigint-migration
  - id: follow-activeflag-converge
    content: user_follow 软删除（active_flag 生成列）+ 表归属收敛：新增 deleted_at 和 active_flag VIRTUAL 列及唯一索引，删除 communication 的 user_baseinfo 重复定义，删除 interaction 的 video_popular 重复定义，VideoVisitAdapter 改为 RPC 调用，删除 user 的冗余 IUserFollowRepo
    status: completed
    dependencies:
      - user-status-softdelete
  - id: idempotency-keys
    content: 幂等性实现：video/comment/liker/relation 四张表新增 idempotency_key 列+唯一索引，实现三段式幂等逻辑（先查→再插→撞1062后再查）
    status: completed
    dependencies:
      - migrate-infra
  - id: relation-stat-table
    content: 新增 user_relation_stat 表：编写迁移文件，在 communication 服务中创建 model 和基础操作
    status: completed
    dependencies:
      - migrate-infra
---

## 用户需求总览

基于 `docs/方案.md`，对 go_zero-tiktok 项目执行以下全套改造：

1. **雪花ID bigint 迁移**：将所有表的 ID 列从 `varchar(64)` 改为 `BIGINT`，数据库存整数，API 层 JSON 序列化为字符串数字，不加前缀
2. **软删除方案落地**：`user_baseinfo` 用 status 状态位（方案 A），`user_follow` 用 active_flag 生成列（方案 B）
3. **表归属收敛**：删除 communication 中 `user_baseinfo` 的重复定义、删除 interaction 中 `video_popular` 的重复定义，跨服务访问改为 RPC 批量查询
4. **幂等性实现**：video、interaction_action、interaction_comment、user_relation、playback_qos_reports 五张表添加 `idempotency_key` 唯一索引
5. **新增 user_relation_stat 表**
6. **迁移管理方式切换为 golang-migrate**
7. **计数一致性问题暂时保留，不做改动**

## 技术栈与实施策略

### 技术栈

- 框架：go-zero v1.10.1（保持不变）
- ORM：GORM v1.31.1 + MySQL driver v1.6.0（保持不变）
- 迁移工具：golang-migrate/migrate v4（新增）
- 雪花 ID：bwmarrin/snowflake v0.3.0（已有）
- 语言：Go 1.26.1

### 实施总体策略

按依赖关系分 6 个阶段顺序执行，每阶段完成一个独立可验证的里程碑。所有数据库变更通过 golang-migrate 创建的迁移文件管理，采用 `up/down` 双向迁移保证可回滚。

#### 阶段 1：golang-migrate 基础设施

- 新增 `migrations/` 目录，将现有 `sql/schema.sql` 拆解为初始迁移
- 在 `pkg/` 下封装 migrate 启动器，各服务启动时自动执行迁移

#### 阶段 2：雪花 ID bigint 全量迁移

- 数据库层：所有 ID 列 `varchar(64)` → `BIGINT`，同步修改主键和外键
- Model 层：所有 GORM 模型 ID 字段 `string` → `int64`
- Proto 层：所有 `string` ID 字段 → `int64`
- 契约层：`pkg/contract/types.go` 中所有 ID 字段 `string` → `int64`，JSON tag 加 `,string` 选项保证序列化为字符串
- 业务逻辑层：移除所有 `fmt.Sprintf` 类型 ID 转换，改为直接使用 int64

#### 阶段 3：user_baseinfo 软删除（status 方案）

- 新增 `status TINYINT NOT NULL DEFAULT 1`
- 唯一索引从 `UNIQUE(username)` 改为 `UNIQUE(username, status)`
- 查询统一加 `status = 1` 过滤
- 删除操作改为 `UPDATE status = 0`

#### 阶段 4：user_follow 软删除 + 表归属收敛

- user_follow 新增 `deleted_at` 和 `active_flag` VIRTUAL 生成列，唯一索引 `(follower_id, followed_id, active_flag)`
- 删除 communication 中 `user_baseinfo` 的 model/mysql 文件，改为通过 user RPC 获取
- 删除 interaction 中 `video_popular` 的 model/mysql 文件，`VideoVisitAdapter` 改为调用 video RPC
- 删除 user 服务中冗余的 `IUserFollowRepo` 接口定义

#### 阶段 5：幂等性实现

- 五张表各新增 `idempotency_key VARCHAR(128)` 可空列 + 组合唯一索引
- 实现 GCFeed 风格的三段式幂等逻辑：先查 → 再插 → 撞 1062 后再查

#### 阶段 6：user_relation_stat 新表

- 新建 `user_relation_stat` 表，含 `user_id`、粉丝数、关注数等统计字段
- 在 communication 服务中创建 model 和基础 CRUD