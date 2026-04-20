package dal

import (
	"context"

	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type Repositories struct {
	User       *UserBaseinfoRepo
	Video      *VideoBaseinfoRepo
	Popular    *VideoPopularRepo
	Comment    *CommentRepo
	VideoLiker *VideoLikerRepo
	UserFollow *UserFollowRepo
}

func NewRepositories(db *gorm.DB, rdb *redis.Redis) *Repositories {
	return &Repositories{
		User:       NewUserBaseinfoRepo(db),
		Video:      NewVideoBaseinfoRepo(db),
		Popular:    NewVideoPopularRepo(db, rdb),
		Comment:    NewCommentRepo(db),
		VideoLiker: NewVideoLikerRepo(db),
		UserFollow: NewUserFollowRepo(db),
	}
}

type UserBaseinfoRepo struct {
	db *gorm.DB
}

func NewUserBaseinfoRepo(db *gorm.DB) *UserBaseinfoRepo {
	return &UserBaseinfoRepo{db: db}
}

func (r *UserBaseinfoRepo) CreateUser(ctx context.Context, user *types.UserBaseinfo) error {
	return CreateUser(ctx, user)
}

func (r *UserBaseinfoRepo) GetUserByID(ctx context.Context, userID string) (*types.UserBaseinfo, error) {
	return GetUserByID(ctx, userID)
}

func (r *UserBaseinfoRepo) GetUserByUsername(ctx context.Context, username string) (*types.UserBaseinfo, error) {
	return GetUserByUsername(ctx, username)
}

func (r *UserBaseinfoRepo) UpdateUserPhotoByID(ctx context.Context, userID string, photoURL string) error {
	return UpdateUserPhotoByID(ctx, userID, photoURL)
}

func (r *UserBaseinfoRepo) GetUsersByIDs(ctx context.Context, userIDs []string) ([]types.UserBaseinfo, error) {
	return GetUsersByIDs(ctx, userIDs)
}

type VideoBaseinfoRepo struct {
	db *gorm.DB
}

func NewVideoBaseinfoRepo(db *gorm.DB) *VideoBaseinfoRepo {
	return &VideoBaseinfoRepo{db: db}
}

func (r *VideoBaseinfoRepo) CreateVideo(ctx context.Context, video *types.VideoBaseinfo) error {
	return CreateVideo(ctx, video)
}

func (r *VideoBaseinfoRepo) SearchVideosByKeyword(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	return SearchVideosByKeyword(ctx, keyword, pageNum, pageSize)
}

func (r *VideoBaseinfoRepo) GetVideosByIDs(ctx context.Context, videoIDs []string) ([]types.VideoBaseinfo, error) {
	return GetVideosByIDs(ctx, videoIDs)
}

func (r *VideoBaseinfoRepo) GetVideosByAuthorID(ctx context.Context, authorID string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	return GetVideosByAuthorID(ctx, authorID, pageNum, pageSize)
}

type VideoPopularRepo struct {
	db  *gorm.DB
	rdb *redis.Redis
}

func NewVideoPopularRepo(db *gorm.DB, rdb *redis.Redis) *VideoPopularRepo {
	return &VideoPopularRepo{db: db, rdb: rdb}
}

func (r *VideoPopularRepo) CreatePopularVideo(ctx context.Context, videoID string) error {
	return CreatePopularVideo(ctx, videoID)
}

func (r *VideoPopularRepo) IncreaseVideoVisitCount(ctx context.Context, videoID string, delta int64) error {
	return IncreaseVideoVisitCount(ctx, videoID, delta)
}

func (r *VideoPopularRepo) UpdateVideoLikeCount(ctx context.Context, videoID string, delta int64) error {
	return UpdateVideoLikeCount(ctx, videoID, delta)
}

func (r *VideoPopularRepo) GetPopularVideoIDsByVisitCount(ctx context.Context, pageNum, pageSize int32) ([]string, int64, error) {
	return GetPopularVideoIDsByVisitCountInZset(ctx, pageNum, pageSize)
}

func (r *VideoPopularRepo) SetPopularVideoToRedis(ctx context.Context, video types.VideoBaseinfo, visitCount int64) error {
	return SetPopularVideoToRedis(ctx, video, visitCount)
}

func (r *VideoPopularRepo) IncrVideoVisitCountInRedis(ctx context.Context, videoID string) error {
	return IncrVideoVisitCountInRedis(ctx, videoID)
}

func (r *VideoPopularRepo) GetVideoVisitCountFromRedis(ctx context.Context, pageSize int, pageNum int) ([]PopularVideoWithHeat, error) {
	return GetVideoVisitCountFromRedis(ctx, pageSize, pageNum)
}

func (r *VideoPopularRepo) GetVideoVisitCountByIDInHash(ctx context.Context, videoIDs []string) ([]map[string]string, error) {
	return GetVideoVisitCountByIDInHash(ctx, videoIDs)
}

type CommentRepo struct {
	db *gorm.DB
}

func NewCommentRepo(db *gorm.DB) *CommentRepo {
	return &CommentRepo{db: db}
}

func (r *CommentRepo) CreateComment(ctx context.Context, comment *types.CommentBaseinfo) error {
	return CreateComment(ctx, comment)
}

func (r *CommentRepo) DeleteCommentByID(ctx context.Context, commentID string) error {
	return DeleteCommentByID(ctx, commentID)
}

func (r *CommentRepo) GetCommentsByVideoID(ctx context.Context, videoID string, pageNumber, pageSize int32) ([]types.CommentBaseinfo, int64, error) {
	return GetCommentsByVideoID(ctx, videoID, pageNumber, pageSize)
}

type VideoLikerRepo struct {
	db *gorm.DB
}

func NewVideoLikerRepo(db *gorm.DB) *VideoLikerRepo {
	return &VideoLikerRepo{db: db}
}

func (r *VideoLikerRepo) LikeVideo(ctx context.Context, userID, videoID string) error {
	return LikeVideo(ctx, userID, videoID)
}

func (r *VideoLikerRepo) CancelLikeVideo(ctx context.Context, userID, videoID string) error {
	return CancelLikeVideo(ctx, userID, videoID)
}

func (r *VideoLikerRepo) GetLikedVideoIDsByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
	return GetLikedVideoIDsByUserID(ctx, userID, pageNumber, pageSize)
}

type UserFollowRepo struct {
	db *gorm.DB
}

func NewUserFollowRepo(db *gorm.DB) *UserFollowRepo {
	return &UserFollowRepo{db: db}
}

func (r *UserFollowRepo) CreateUserFollow(ctx context.Context, followerID, userID string) error {
	return CreateUserFollow(ctx, followerID, userID)
}

func (r *UserFollowRepo) UpdateUserFollowStatus(ctx context.Context, followerID, userID string, status int32) error {
	return UpdateUserFollowStatus(ctx, followerID, userID, status)
}

func (r *UserFollowRepo) GetFollowingISSubriber(ctx context.Context, followerID, userID string) (bool, error) {
	return GetFollowingISSubriber(ctx, followerID, userID)
}

func (r *UserFollowRepo) GetFollowingByFollowerID(ctx context.Context, followerID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	return GetFollowingByFollowerID(ctx, followerID, pageNumber, pageSize)
}

func (r *UserFollowRepo) GetFansByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	return GetFansByUserID(ctx, userID, pageNumber, pageSize)
}

func (r *UserFollowRepo) GetFriendByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	return GetFriendByUserID(ctx, userID, pageNumber, pageSize)
}
