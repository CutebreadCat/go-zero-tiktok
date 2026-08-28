-- 视频播放质量聚合指标独立表，避免 video_stat 过度臃肿。
CREATE TABLE `video_qos_stat` (
  `video_id`        bigint   NOT NULL COMMENT '雪花ID，与 video_stat 1:1',
  `completion_rate` int      NOT NULL DEFAULT 0 COMMENT '完播率 万分比(0-10000)',
  `stall_rate`      int      NOT NULL DEFAULT 0 COMMENT '卡顿率 万分比(0-10000)',
  `error_rate`      int      NOT NULL DEFAULT 0 COMMENT '错误率 万分比(0-10000)',
  `avg_bitrate_kbps` int     NOT NULL DEFAULT 0 COMMENT '平均码率 kbps',
  `avg_buffered_ms` bigint   NOT NULL DEFAULT 0 COMMENT '平均缓冲耗时 ms',
  `avg_stall_count` int      NOT NULL DEFAULT 0 COMMENT '平均卡顿次数',
  `report_count`    bigint   NOT NULL DEFAULT 0 COMMENT '聚合上报样本数',
  `updated_at`      datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '聚合更新时间',
  PRIMARY KEY (`video_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='视频播放质量聚合指标';
