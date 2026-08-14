package cache

import "fmt"

const (
	// VideoLikeUsersKey 点赞某视频的用户集合，key 后缀为 video_id。
	// 用于：1) 幂等去重；2) 计算真实 like_count；3) 同步到 MySQL video_interaction。
	VideoLikeUsersKey = "video:like:users:%d"
	// UserLikeVideosKey 某用户点赞的视频有序集合，key 后缀为 user_id。
	// score = 点赞时间毫秒时间戳，用于按时间倒序分页查询。
	UserLikeVideosKey = "user:like:videos:%d"
	// LikeCountKey 维护各视频当前应展示的 like_count 总值（聚合缓存，加速读）。
	LikeCountKey = "video:like:count"
	// LikeDirtyKey 记录有待 flush 到 MySQL 的视频 ID。
	LikeDirtyKey = "video:like:dirty"

	// VideoFavoriteUsersKey 收藏某视频的用户集合。
	VideoFavoriteUsersKey = "video:favorite:users:%d"
	// UserFavoriteVideosKey 某用户收藏的视频有序集合。
	UserFavoriteVideosKey = "user:favorite:videos:%d"
	// FavoriteCountKey 维护各视频当前 favorite_count 总值。
	FavoriteCountKey = "video:favorite:count"
	// FavoriteDirtyKey 记录有待 flush 的收藏视频 ID。
	FavoriteDirtyKey = "video:favorite:dirty"

	// likeKeyTTLSeconds Redis 点赞/收藏相关 key 的过期时间：30 天。
	likeKeyTTLSeconds = 30 * 24 * 60 * 60
)

func fmtVideoLikeUsersKey(videoID int64) string {
	return fmt.Sprintf(VideoLikeUsersKey, videoID)
}

func fmtUserLikeVideosKey(userID int64) string {
	return fmt.Sprintf(UserLikeVideosKey, userID)
}

func fmtVideoFavoriteUsersKey(videoID int64) string {
	return fmt.Sprintf(VideoFavoriteUsersKey, videoID)
}

func fmtUserFavoriteVideosKey(userID int64) string {
	return fmt.Sprintf(UserFavoriteVideosKey, userID)
}

func toAnySlice(ss []string) []any {
	result := make([]any, len(ss))
	for i, s := range ss {
		result[i] = s
	}
	return result
}
