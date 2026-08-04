CREATE TABLE `user_baseinfo` (
  `user_id` varchar(64) NOT NULL,
  `username` varchar(64) NOT NULL UNIQUE,
  `password` varchar(255) NOT NULL,
  `photo_url` varchar(255) NOT NULL,
  `created_at` longtext,
  `updated_at` longtext,
  `deleted_at` longtext,
  PRIMARY KEY (`user_id`)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `user_follow` (
  `follower_id` varchar(64) NOT NULL,
  `user_id` varchar(64) NOT NULL,
  `status` int NOT NULL DEFAULT 0,
  PRIMARY KEY (`follower_id`, `user_id`)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `video_baseinfo` (
  `video_id` varchar(64) NOT NULL,
  `author_id` varchar(64) NOT NULL,
  `video_url` varchar(512) NOT NULL,
  `cover_url` varchar(512) DEFAULT NULL,
  `title` varchar(255) NOT NULL,
  `description` varchar(512) DEFAULT NULL,
  `created_at` longtext,
  `updated_at` longtext,
  `deleted_at` longtext,
  PRIMARY KEY (`video_id`),
  INDEX `idx_author_id` (`author_id`)
)  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `video_popular` (
  `video_id` varchar(64) NOT NULL,
  `visit_count` bigint NOT NULL DEFAULT 0,
  `like_count` bigint NOT NULL DEFAULT 0,
  `comment_count` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`video_id`)
)  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `video_liker` (
  `user_id` varchar(64) NOT NULL,
  `video_id` varchar(64) NOT NULL,
  PRIMARY KEY (`user_id`, `video_id`)
)  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `comment_baseinfo` (
  `comment_id` varchar(64) NOT NULL,
  `user_id` varchar(64) NOT NULL,
  `video_id` varchar(64) NOT NULL,
  `content` varchar(1024) NOT NULL,
  `created_at` longtext,
  `updated_at` longtext,
  `deleted_at` longtext,
  `like_count` int NOT NULL DEFAULT 0,
  `parent_comment_id` char(64) NOT NULL DEFAULT '',
  PRIMARY KEY (`comment_id`),
  INDEX `idx_video_id` (`video_id`)
)  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `comment_liker` (
  `user_id` varchar(64) NOT NULL,
  `comment_id` varchar(64) NOT NULL,
  PRIMARY KEY (`user_id`, `comment_id`)
)  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `user_mfa` (
  `user_id` varchar(64) NOT NULL,
  `mfa_secret` varchar(255) NOT NULL,
  `mfa_enabled` tinyint(1) NOT NULL DEFAULT 0,
  `password_hash` varchar(255) NOT NULL,
  `mfa_pending_secret` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`user_id`)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


