-- 000001 初始化数据库结构（重构后目标形态）
-- 设计依据：docs/方案.md
-- 1. 雪花ID bigint：所有 ID varchar(64) -> BIGINT
-- 2. user_baseinfo: status 状态位软删（方案A），UNIQUE(username, status)
-- 3. user_follow: active_flag 生成列软删（方案B），保留 deleted_at，自增主键
-- 4. video_popular 改名 video_stat（归属 video 服务）
-- 5. 幂等键 idempotency_key 进唯一索引（GCFeed 模式）
-- 6. 新增 user_relation_stat（communication 服务）
-- 7. 单库，时间戳统一 datetime(3)
-- 8. user_mfa 合并入 user_baseinfo（1:1 关系，删除冗余 password_hash）

CREATE TABLE `user_baseinfo` (
  `user_id`            bigint       NOT NULL COMMENT '雪花ID',
  `username`           varchar(64)  NOT NULL COMMENT '用户名',
  `password`           varchar(64)  NOT NULL COMMENT '密码SHA-256 hex(64字符)',
  `photo_url`          varchar(255) NOT NULL COMMENT '头像URL',
  `mfa_secret`         varchar(64)  DEFAULT NULL COMMENT 'TOTP base32密钥',
  `mfa_enabled`        tinyint(1)   NOT NULL DEFAULT 0 COMMENT 'MFA是否启用',
  `mfa_pending_secret` varchar(64)  DEFAULT NULL COMMENT '待绑定TOTP密钥',
  `status`             tinyint      NOT NULL DEFAULT 1 COMMENT '1正常 0已删 2封禁',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`user_id`),
  UNIQUE KEY `uk_username_status` (`username`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户基础信息(含MFA)';

CREATE TABLE `user_follow` (
  `id`              bigint       NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `follower_id`     bigint       NOT NULL COMMENT '关注者雪花ID',
  `user_id`         bigint       NOT NULL COMMENT '被关注者雪花ID',
  `idempotency_key` varchar(64) DEFAULT NULL COMMENT '幂等键(UUID 36字符)',
  `deleted_at`      datetime(3)  DEFAULT NULL COMMENT '取关时间 NULL=关注中',
  `active_flag`     tinyint      GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, 1, NULL)) VIRTUAL COMMENT '生成列:活跃标记',
  `created_at`      datetime(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`      datetime(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_active_relation` (`follower_id`, `user_id`, `active_flag`),
  UNIQUE KEY `uk_follower_idempotency` (`follower_id`, `idempotency_key`),
  KEY `idx_follow_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='关注关系';

CREATE TABLE `video_baseinfo` (
  `video_id`         bigint       NOT NULL COMMENT '雪花ID',
  `author_id`        bigint       NOT NULL COMMENT '作者雪花ID',
  `video_object_key` varchar(255) NOT NULL COMMENT '视频 OSS object key',
  `cover_object_key` varchar(255) DEFAULT NULL COMMENT '封面 OSS object key',
  `title`            varchar(128) NOT NULL COMMENT '视频标题',
  `description`      varchar(255) DEFAULT NULL COMMENT '视频描述',
  `idempotency_key`  varchar(64) DEFAULT NULL COMMENT '幂等键(UUID 36字符)',
  `deleted_at`       datetime(3)  DEFAULT NULL COMMENT '软删时间',
  `created_at`       datetime(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`       datetime(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`video_id`),
  UNIQUE KEY `uk_author_idempotency` (`author_id`, `idempotency_key`),
  KEY `idx_author_id` (`author_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='视频基础信息';

CREATE TABLE `video_stat` (
  `video_id`        bigint NOT NULL COMMENT '雪花ID',
  `visit_count`     bigint NOT NULL DEFAULT 0,
  `like_count`      bigint NOT NULL DEFAULT 0,
  `comment_count`   bigint NOT NULL DEFAULT 0,
  `favorite_count`  bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`video_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='视频统计(原video_popular)';

CREATE TABLE `video_liker` (
  `user_id`         bigint       NOT NULL COMMENT '点赞用户雪花ID',
  `video_id`        bigint       NOT NULL COMMENT '视频雪花ID',
  `idempotency_key` varchar(64) DEFAULT NULL COMMENT '幂等键(UUID 36字符)',
  `created_at`      datetime(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`user_id`, `video_id`),
  UNIQUE KEY `uk_user_idempotency` (`user_id`, `idempotency_key`),
  KEY `idx_video_id` (`video_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='视频点赞';

CREATE TABLE `video_favoriter` (
  `user_id`         bigint       NOT NULL COMMENT '收藏用户雪花ID',
  `video_id`        bigint       NOT NULL COMMENT '视频雪花ID',
  `idempotency_key` varchar(64) DEFAULT NULL COMMENT '幂等键(UUID 36字符)',
  `created_at`      datetime(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`user_id`, `video_id`),
  UNIQUE KEY `uk_user_idempotency` (`user_id`, `idempotency_key`),
  KEY `idx_video_id` (`video_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='视频收藏';

CREATE TABLE `comment_baseinfo` (
  `comment_id`      bigint        NOT NULL COMMENT '雪花ID',
  `user_id`         bigint        NOT NULL COMMENT '评论者雪花ID',
  `video_id`        bigint        NOT NULL COMMENT '视频雪花ID',
  `content`         varchar(512) NOT NULL COMMENT '评论内容',
  `like_count`      bigint        NOT NULL DEFAULT 0 COMMENT '点赞数',
  `parent_comment_id` bigint      NOT NULL DEFAULT 0 COMMENT '父评论雪花ID 0=根评论',
  `idempotency_key` varchar(64)  DEFAULT NULL COMMENT '幂等键(UUID 36字符)',
  `deleted_at`      datetime(3)   DEFAULT NULL COMMENT '软删时间',
  `created_at`      datetime(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`      datetime(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`comment_id`),
  UNIQUE KEY `uk_user_idempotency` (`user_id`, `idempotency_key`),
  KEY `idx_video_id` (`video_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论基础信息';

CREATE TABLE `comment_liker` (
  `user_id`         bigint       NOT NULL COMMENT '点赞用户雪花ID',
  `comment_id`      bigint       NOT NULL COMMENT '评论雪花ID',
  `idempotency_key` varchar(64) DEFAULT NULL COMMENT '幂等键(UUID 36字符)',
  `created_at`      datetime(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`user_id`, `comment_id`),
  UNIQUE KEY `uk_user_idempotency` (`user_id`, `idempotency_key`),
  KEY `idx_comment_id` (`comment_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论点赞';

CREATE TABLE `playback_qos_reports` (
  `id`              bigint       NOT NULL AUTO_INCREMENT,
  `user_id`         bigint       NOT NULL COMMENT '用户雪花ID',
  `video_id`        bigint       NOT NULL COMMENT '视频雪花ID',
  `report_data`     json         DEFAULT NULL COMMENT '播放质量上报数据',
  `idempotency_key` varchar(64) DEFAULT NULL COMMENT '幂等键(UUID 36字符)',
  `created_at`      datetime(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_idempotency` (`user_id`, `idempotency_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='播放质量上报';

CREATE TABLE `user_relation_stat` (
  `user_id`         bigint      NOT NULL COMMENT '用户雪花ID',
  `follower_count`  bigint      NOT NULL DEFAULT 0 COMMENT '粉丝数',
  `following_count` bigint      NOT NULL DEFAULT 0 COMMENT '关注数',
  `friend_count`    bigint      NOT NULL DEFAULT 0 COMMENT '互关数',
  `created_at`      datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`      datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户关系统计';

