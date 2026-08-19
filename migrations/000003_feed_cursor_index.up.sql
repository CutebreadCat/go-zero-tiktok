-- Feed 复合游标分页：WHERE (created_at, video_id) < (?, ?) 需要复合索引避免全表扫描。
ALTER TABLE `video_baseinfo` ADD KEY `idx_created_at_video_id` (`created_at`, `video_id`);
