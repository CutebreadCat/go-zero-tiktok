-- 视频观看事件流水：曝光、播放、完播、跳过
CREATE TABLE `video_view_events` (
  `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `user_id` BIGINT NOT NULL COMMENT '上报用户',
  `video_id` BIGINT NOT NULL COMMENT '视频 ID',
  `scene` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'feed 场景',
  `request_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '一次请求标识',
  `event_type` VARCHAR(32) NOT NULL COMMENT 'exposed|play|complete|skip',
  `watch_ms` BIGINT NOT NULL DEFAULT 0 COMMENT '观看时长毫秒',
  `completed` TINYINT NOT NULL DEFAULT 0 COMMENT '是否完播',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY `idx_user_video` (`user_id`, `video_id`),
  KEY `idx_video_event` (`video_id`, `event_type`, `created_at`),
  KEY `idx_request_id` (`request_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='视频观看事件流水';

-- 用户视频曝光聚合：用于快速判断用户是否看过某视频
CREATE TABLE `user_video_exposures` (
  `user_id` BIGINT NOT NULL,
  `video_id` BIGINT NOT NULL,
  `first_exposed_at` DATETIME NOT NULL,
  `last_exposed_at` DATETIME NOT NULL,
  `exposure_count` INT NOT NULL DEFAULT 1,
  `last_scene` VARCHAR(32) NOT NULL DEFAULT '',
  PRIMARY KEY (`user_id`, `video_id`),
  KEY `idx_user_time` (`user_id`, `last_exposed_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户视频曝光聚合';
