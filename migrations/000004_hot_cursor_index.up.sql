-- Hot 场景复合游标分页：WHERE (visit_count, video_id) < (?, ?) 需要复合索引避免全表扫描。
ALTER TABLE `video_stat` ADD KEY `idx_visit_count_video_id` (`visit_count`, `video_id`);
