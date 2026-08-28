// Package testhelpers 提供单元测试共享设施：
// 基于 glebarez/sqlite 的临时文件数据库，无需 MySQL/cgo，可离线运行纯逻辑单测。
//
// 建表采用与生产迁移（migrations/000001_init_schema.up.sql）等价的 DDL，
// 仅剥离 MySQL 专有子句（ENGINE/CHARSET/COMMENT/AUTO_INCREMENT/ON UPDATE）。
// 这样可保证：
//  1. 时间戳列 DATETIME 亲和，驱动回读时自动转为 time.Time（与生产一致）；
//  2. user_follow 的 active_flag 生成列真实生效，软删恢复语义与生产一致；
//  3. 唯一索引/外键语义与生产对齐，幂等、游标测试可靠。
package testhelpers

import (
	"os"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewTestDB 创建一个隔离的临时文件 SQLite 数据库，并按生产 DDL 建表。
// 使用临时文件（而非 :memory: 共享缓存）可彻底避免连接池回收导致整库丢失、
// 后续查询报 "no such table" 的抖动问题；测试结束自动清理文件。
func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	f, err := os.CreateTemp("", "tiktok-test-*.db")
	if err != nil {
		t.Fatalf("create temp db file failed: %v", err)
	}
	name := f.Name()
	_ = f.Close()
	t.Cleanup(func() { _ = os.Remove(name) })

	db, err := gorm.Open(sqlite.Open(name), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}

	for _, ddl := range schemaDDL() {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatalf("create table failed: %v\nddl: %s", err, ddl)
		}
	}

	return db
}

// schemaDDL 返回各表的生产等价建表语句（已适配 SQLite 语法）。
func schemaDDL() []string {
	return []string{
		`CREATE TABLE user_baseinfo (
			user_id            bigint       NOT NULL,
			username           varchar(64)  NOT NULL,
			password           varchar(64)  NOT NULL,
			photo_url          varchar(255) NOT NULL,
			mfa_secret         varchar(64)  DEFAULT NULL,
			mfa_enabled        tinyint      NOT NULL DEFAULT 0,
			mfa_pending_secret varchar(64)  DEFAULT NULL,
			status             tinyint      NOT NULL DEFAULT 1,
			created_at         datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at         datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id),
			UNIQUE (username, status)
		)`,
		`CREATE TABLE user_follow (
			id              INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			follower_id     bigint   NOT NULL,
			user_id         bigint   NOT NULL,
			idempotency_key varchar(64) DEFAULT NULL,
			deleted_at      datetime DEFAULT NULL,
			active_flag     tinyint  GENERATED ALWAYS AS (CASE WHEN deleted_at IS NULL THEN 1 ELSE NULL END) VIRTUAL,
			created_at      datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at      datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (follower_id, user_id, active_flag),
			UNIQUE (follower_id, idempotency_key)
		)`,
		`CREATE INDEX idx_follow_user ON user_follow(user_id)`,
		`CREATE TABLE video_baseinfo (
			video_id         bigint       NOT NULL,
			author_id        bigint       NOT NULL,
			video_object_key varchar(255) NOT NULL,
			cover_object_key varchar(255) DEFAULT NULL,
			title            varchar(128) NOT NULL,
			description      varchar(255) DEFAULT NULL,
			idempotency_key  varchar(64)  DEFAULT NULL,
			deleted_at       datetime     DEFAULT NULL,
			created_at       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (video_id),
			UNIQUE (author_id, idempotency_key)
		)`,
		`CREATE INDEX idx_author_id ON video_baseinfo(author_id)`,
		`CREATE TABLE video_stat (
			video_id        bigint NOT NULL,
			visit_count     bigint NOT NULL DEFAULT 0,
			like_count      bigint NOT NULL DEFAULT 0,
			comment_count   bigint NOT NULL DEFAULT 0,
			favorite_count  bigint NOT NULL DEFAULT 0,
			PRIMARY KEY (video_id)
		)`,
		`CREATE TABLE video_interaction (
			id              INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			user_id         bigint       NOT NULL,
			video_id        bigint       NOT NULL,
			action_type     tinyint      NOT NULL DEFAULT 1,
			idempotency_key varchar(64)  DEFAULT NULL,
			created_at      datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (user_id, video_id, action_type),
			UNIQUE (user_id, action_type, idempotency_key)
		)`,
		`CREATE INDEX idx_video_id ON video_interaction(video_id)`,
		`CREATE TABLE comment_baseinfo (
			comment_id       bigint        NOT NULL,
			user_id          bigint        NOT NULL,
			video_id         bigint        NOT NULL,
			content          varchar(512)  NOT NULL,
			like_count       bigint        NOT NULL DEFAULT 0,
			parent_comment_id bigint       NOT NULL DEFAULT 0,
			idempotency_key  varchar(64)   DEFAULT NULL,
			deleted_at       datetime      DEFAULT NULL,
			created_at       datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at       datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (comment_id),
			UNIQUE (user_id, idempotency_key)
		)`,
		`CREATE INDEX idx_comment_video_id ON comment_baseinfo(video_id)`,
		`CREATE INDEX idx_comment_user_id ON comment_baseinfo(user_id)`,
		`CREATE TABLE comment_liker (
			user_id         bigint       NOT NULL,
			comment_id      bigint       NOT NULL,
			idempotency_key varchar(64)  DEFAULT NULL,
			created_at      datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, comment_id),
			UNIQUE (user_id, idempotency_key)
		)`,
		`CREATE INDEX idx_comment_id ON comment_liker(comment_id)`,
		`CREATE TABLE playback_qos_reports (
			id              INTEGER  NOT NULL PRIMARY KEY AUTOINCREMENT,
			user_id         bigint   NOT NULL,
			video_id        bigint   NOT NULL,
			report_data     text     DEFAULT NULL,
			idempotency_key varchar(64) DEFAULT NULL,
			created_at      datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (user_id, idempotency_key)
		)`,
		`CREATE TABLE video_qos_stat (
			video_id         bigint   NOT NULL PRIMARY KEY,
			completion_rate  int      NOT NULL DEFAULT 0,
			stall_rate       int      NOT NULL DEFAULT 0,
			error_rate       int      NOT NULL DEFAULT 0,
			avg_bitrate_kbps int      NOT NULL DEFAULT 0,
			avg_buffered_ms  bigint   NOT NULL DEFAULT 0,
			avg_stall_count  int      NOT NULL DEFAULT 0,
			report_count     bigint   NOT NULL DEFAULT 0,
			updated_at       datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE user_relation_stat (
			user_id         bigint   NOT NULL,
			follower_count  bigint   NOT NULL DEFAULT 0,
			following_count bigint   NOT NULL DEFAULT 0,
			friend_count    bigint   NOT NULL DEFAULT 0,
			created_at      datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at      datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id)
		)`,
		`CREATE TABLE messages (
			id                  INTEGER      NOT NULL PRIMARY KEY AUTOINCREMENT,
			message_id          BIGINT       NOT NULL,
			receiver_id         BIGINT       NOT NULL,
			type                VARCHAR(16)  NOT NULL,
			title               VARCHAR(128) NOT NULL,
			content             VARCHAR(1024) NOT NULL,
			event_id            VARCHAR(64)  DEFAULT NULL,
			sender_id           BIGINT       NOT NULL DEFAULT 0,
			sender_nickname     VARCHAR(64)  NOT NULL DEFAULT '',
			sender_avatar_url   VARCHAR(512) NOT NULL DEFAULT '',
			target_id           BIGINT       NOT NULL DEFAULT 0,
			target_type         VARCHAR(32)  NOT NULL DEFAULT '',
			is_read             TINYINT      NOT NULL DEFAULT 0,
			created_at          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
			read_at             DATETIME     DEFAULT NULL,
			UNIQUE (message_id),
			UNIQUE (receiver_id, event_id)
		)`,
		`CREATE INDEX idx_receiver_read_created ON messages(receiver_id, is_read, created_at)`,
	}
}
