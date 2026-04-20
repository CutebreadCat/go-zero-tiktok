package dal

import (
	"context"

	mysqlrepo "go_zero-tiktok/internal/dal/repository/mysql"
	redisrepo "go_zero-tiktok/internal/dal/repository/redis"
	videopopulartable "go_zero-tiktok/internal/dal/tables/video_popular"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type Repositories struct {
	Mysql *MysqlRepositories
	Redis *RedisRepositories

	// Flat facades are kept for backward compatibility with existing business code.
	User       *UserBaseinfoRepo
	Video      *VideoBaseinfoRepo
	Popular    *VideoPopularRepo
	Comment    *CommentRepo
	VideoLiker *VideoLikerRepo
	UserFollow *UserFollowRepo
}

type MysqlRepositories struct {
	User       *mysqlrepo.UserBaseinfoRepo
	Video      *mysqlrepo.VideoBaseinfoRepo
	Popular    *mysqlrepo.VideoPopularRepo
	Comment    *mysqlrepo.CommentRepo
	VideoLiker *mysqlrepo.VideoLikerRepo
	UserFollow *mysqlrepo.UserFollowRepo
}

type RedisRepositories struct {
	Popular    *redisrepo.VideoPopularRepo
	VideoLiker *redisrepo.VideoLikerRepo
}

func NewRepositories(db *gorm.DB, rdb *redis.Redis) *Repositories {
	mysqlRepos := &MysqlRepositories{
		User:       mysqlrepo.NewUserBaseinfoRepo(db),
		Video:      mysqlrepo.NewVideoBaseinfoRepo(db),
		Popular:    mysqlrepo.NewVideoPopularRepo(db),
		Comment:    mysqlrepo.NewCommentRepo(db),
		VideoLiker: mysqlrepo.NewVideoLikerRepo(db),
		UserFollow: mysqlrepo.NewUserFollowRepo(db),
	}

	redisRepos := &RedisRepositories{
		Popular:    redisrepo.NewVideoPopularRepo(rdb),
		VideoLiker: redisrepo.NewVideoLikerRepo(rdb),
	}

	return &Repositories{
		Mysql: mysqlRepos,
		Redis: redisRepos,

		User:       &UserBaseinfoRepo{mysql: mysqlRepos.User},
		Video:      &VideoBaseinfoRepo{mysql: mysqlRepos.Video},
		Popular:    &VideoPopularRepo{mysql: mysqlRepos.Popular, redis: redisRepos.Popular},
		Comment:    &CommentRepo{mysql: mysqlRepos.Comment},
		VideoLiker: &VideoLikerRepo{mysql: mysqlRepos.VideoLiker, redis: redisRepos.VideoLiker},
		UserFollow: &UserFollowRepo{mysql: mysqlRepos.UserFollow},
	}
}

type UserBaseinfoRepo struct {
	mysql *mysqlrepo.UserBaseinfoRepo
}

func NewUserBaseinfoRepo(db *gorm.DB) *UserBaseinfoRepo {
	return &UserBaseinfoRepo{mysql: mysqlrepo.NewUserBaseinfoRepo(db)}
}

func (r *UserBaseinfoRepo) CreateUser(ctx context.Context, user *types.UserBaseinfo) error {
	return r.mysql.CreateUser(ctx, user)
}

func (r *UserBaseinfoRepo) GetUserByID(ctx context.Context, userID string) (*types.UserBaseinfo, error) {
	return r.mysql.GetUserByID(ctx, userID)
}

func (r *UserBaseinfoRepo) GetUserByUsername(ctx context.Context, username string) (*types.UserBaseinfo, error) {
	return r.mysql.GetUserByUsername(ctx, username)
}

func (r *UserBaseinfoRepo) UpdateUserPhotoByID(ctx context.Context, userID string, photoURL string) error {
	return r.mysql.UpdateUserPhotoByID(ctx, userID, photoURL)
}

func (r *UserBaseinfoRepo) GetUsersByIDs(ctx context.Context, userIDs []string) ([]types.UserBaseinfo, error) {
	return r.mysql.GetUsersByIDs(ctx, userIDs)
}

type VideoBaseinfoRepo struct {
	mysql *mysqlrepo.VideoBaseinfoRepo
}

func NewVideoBaseinfoRepo(db *gorm.DB) *VideoBaseinfoRepo {
	return &VideoBaseinfoRepo{mysql: mysqlrepo.NewVideoBaseinfoRepo(db)}
}

func (r *VideoBaseinfoRepo) CreateVideo(ctx context.Context, video *types.VideoBaseinfo) error {
	return r.mysql.CreateVideo(ctx, video)
}

func (r *VideoBaseinfoRepo) SearchVideosByKeyword(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	return r.mysql.SearchVideosByKeyword(ctx, keyword, pageNum, pageSize)
}

func (r *VideoBaseinfoRepo) GetVideosByIDs(ctx context.Context, videoIDs []string) ([]types.VideoBaseinfo, error) {
	return r.mysql.GetVideosByIDs(ctx, videoIDs)
}

func (r *VideoBaseinfoRepo) GetVideosByAuthorID(ctx context.Context, authorID string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	return r.mysql.GetVideosByAuthorID(ctx, authorID, pageNum, pageSize)
}

type VideoPopularRepo struct {
	mysql *mysqlrepo.VideoPopularRepo
	redis *redisrepo.VideoPopularRepo
}

func NewVideoPopularRepo(db *gorm.DB, rdb *redis.Redis) *VideoPopularRepo {
	return &VideoPopularRepo{
		mysql: mysqlrepo.NewVideoPopularRepo(db),
		redis: redisrepo.NewVideoPopularRepo(rdb),
	}
}

func (r *VideoPopularRepo) CreatePopularVideo(ctx context.Context, videoID string) error {
	return r.mysql.CreatePopularVideo(ctx, videoID)
}

func (r *VideoPopularRepo) IncreaseVideoVisitCount(ctx context.Context, videoID string, delta int64) error {
	return r.mysql.IncreaseVideoVisitCount(ctx, videoID, delta)
}

func (r *VideoPopularRepo) UpdateVideoLikeCount(ctx context.Context, videoID string, delta int64) error {
	return r.mysql.UpdateVideoLikeCount(ctx, videoID, delta)
}

func (r *VideoPopularRepo) GetPopularVideoIDsByVisitCount(ctx context.Context, pageNum, pageSize int32) ([]string, int64, error) {
	return r.mysql.GetPopularVideoIDsByVisitCount(ctx, pageNum, pageSize)
}

func (r *VideoPopularRepo) SetPopularVideoToRedis(ctx context.Context, video types.VideoBaseinfo, visitCount int64) error {
	return r.redis.SetPopularVideoToRedis(ctx, video, visitCount)
}

func (r *VideoPopularRepo) IncrVideoVisitCountInRedis(ctx context.Context, videoID string) error {
	return r.redis.IncrVideoVisitCountInRedis(ctx, videoID)
}

func (r *VideoPopularRepo) GetVideoVisitCountFromRedis(ctx context.Context, pageSize int, pageNum int) ([]videopopulartable.PopularVideoWithHeat, error) {
	return r.redis.GetVideoVisitCountFromRedis(ctx, pageSize, pageNum)
}

func (r *VideoPopularRepo) GetVideoVisitCountByIDInHash(ctx context.Context, videoIDs []string) ([]map[string]string, error) {
	return r.redis.GetVideoVisitCountByIDInHash(ctx, videoIDs)
}

type CommentRepo struct {
	mysql *mysqlrepo.CommentRepo
}

func NewCommentRepo(db *gorm.DB) *CommentRepo {
	return &CommentRepo{mysql: mysqlrepo.NewCommentRepo(db)}
}

func (r *CommentRepo) CreateComment(ctx context.Context, comment *types.CommentBaseinfo) error {
	return r.mysql.CreateComment(ctx, comment)
}

func (r *CommentRepo) DeleteCommentByID(ctx context.Context, commentID string) error {
	return r.mysql.DeleteCommentByID(ctx, commentID)
}

func (r *CommentRepo) GetCommentsByVideoID(ctx context.Context, videoID string, pageNumber, pageSize int32) ([]types.CommentBaseinfo, int64, error) {
	return r.mysql.GetCommentsByVideoID(ctx, videoID, pageNumber, pageSize)
}

type VideoLikerRepo struct {
	mysql *mysqlrepo.VideoLikerRepo
	redis *redisrepo.VideoLikerRepo
}

func NewVideoLikerRepo(db *gorm.DB, rdb *redis.Redis) *VideoLikerRepo {
	return &VideoLikerRepo{
		mysql: mysqlrepo.NewVideoLikerRepo(db),
		redis: redisrepo.NewVideoLikerRepo(rdb),
	}
}

func (r *VideoLikerRepo) LikeVideo(ctx context.Context, userID, videoID string) error {
	if err := r.mysql.LikeVideo(ctx, userID, videoID); err != nil {
		return err
	}

	// MySQL is source of truth; cache sync failure should not block main flow.
	if err := r.redis.AddVideoLike(ctx, userID, videoID); err != nil {
		logx.WithContext(ctx).Errorf("sync like to redis failed: %v", err)
	}

	return nil
}

func (r *VideoLikerRepo) CancelLikeVideo(ctx context.Context, userID, videoID string) error {
	if err := r.mysql.CancelLikeVideo(ctx, userID, videoID); err != nil {
		return err
	}

	// MySQL delete succeeded; cache cleanup failure is tolerated.
	if err := r.redis.RemoveVideoLike(ctx, userID, videoID); err != nil {
		logx.WithContext(ctx).Errorf("sync unlike to redis failed: %v", err)
	}

	return nil
}

func (r *VideoLikerRepo) GetLikedVideoIDsByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
	return r.mysql.GetLikedVideoIDsByUserID(ctx, userID, pageNumber, pageSize)
}

type UserFollowRepo struct {
	mysql *mysqlrepo.UserFollowRepo
}

func NewUserFollowRepo(db *gorm.DB) *UserFollowRepo {
	return &UserFollowRepo{mysql: mysqlrepo.NewUserFollowRepo(db)}
}

func (r *UserFollowRepo) CreateUserFollow(ctx context.Context, followerID, userID string) error {
	return r.mysql.CreateUserFollow(ctx, followerID, userID)
}

func (r *UserFollowRepo) UpdateUserFollowStatus(ctx context.Context, followerID, userID string, status int32) error {
	return r.mysql.UpdateUserFollowStatus(ctx, followerID, userID, status)
}

func (r *UserFollowRepo) GetFollowingISSubriber(ctx context.Context, followerID, userID string) (bool, error) {
	return r.mysql.GetFollowingISSubriber(ctx, followerID, userID)
}

func (r *UserFollowRepo) GetFollowingByFollowerID(ctx context.Context, followerID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	return r.mysql.GetFollowingByFollowerID(ctx, followerID, pageNumber, pageSize)
}

func (r *UserFollowRepo) GetFansByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	return r.mysql.GetFansByUserID(ctx, userID, pageNumber, pageSize)
}

func (r *UserFollowRepo) GetFriendByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	return r.mysql.GetFriendByUserID(ctx, userID, pageNumber, pageSize)
}
