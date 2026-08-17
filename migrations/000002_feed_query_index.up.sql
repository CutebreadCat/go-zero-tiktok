-- Feed 查询优化：兜底查询 GetVideoByLastTime 使用 WHERE created_at < ?，
-- 无索引时全表扫描 + filesort，补充 created_at 索引。
ALTER TABLE `video_baseinfo` ADD KEY `idx_created_at` (`created_at`);
