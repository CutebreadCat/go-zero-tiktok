package mysql

import (
	"context"

	commenttable "go_zero-tiktok/internal/dal/tables/comment_baseinfo"
	userbasetable "go_zero-tiktok/internal/dal/tables/user_baseinfo"
	userfollowtable "go_zero-tiktok/internal/dal/tables/user_follow"
	videobasetable "go_zero-tiktok/internal/dal/tables/video_baseinfo"
	videolikertable "go_zero-tiktok/internal/dal/tables/video_liker"
	videopopulartable "go_zero-tiktok/internal/dal/tables/video_popular"
	"go_zero-tiktok/internal/types"

	"gorm.io/gorm"
)

type UserBaseinfoRepo struct {
	db *gorm.DB
}

func NewUserBaseinfoRepo(db *gorm.DB) *UserBaseinfoRepo {
	return &UserBaseinfoRepo{db: db}
}

func (r *UserBaseinfoRepo) CreateUser(ctx context.Context, user *types.UserBaseinfo) error {
	return userbasetable.CreateUser(ctx, r.db, user)
}

func (r *UserBaseinfoRepo) GetUserByID(ctx context.Context, userID string) (*types.UserBaseinfo, error) {
	return userbasetable.GetUserByID(ctx, r.db, userID)
}

func (r *UserBaseinfoRepo) GetUserByUsername(ctx context.Context, username string) (*types.UserBaseinfo, error) {
	return userbasetable.GetUserByUsername(ctx, r.db, username)
}

func (r *UserBaseinfoRepo) UpdateUserPhotoByID(ctx context.Context, userID string, photoURL string) error {
	return userbasetable.UpdateUserPhotoByID(ctx, r.db, userID, photoURL)
}

func (r *UserBaseinfoRepo) GetUsersByIDs(ctx context.Context, userIDs []string) ([]types.UserBaseinfo, error) {
	return userbasetable.GetUsersByIDs(ctx, r.db, userIDs)
}

type VideoBaseinfoRepo struct {
	db *gorm.DB
}

func NewVideoBaseinfoRepo(db *gorm.DB) *VideoBaseinfoRepo {
	return &VideoBaseinfoRepo{db: db}
}

func (r *VideoBaseinfoRepo) CreateVideo(ctx context.Context, video *types.VideoBaseinfo) error {
	return videobasetable.CreateVideo(ctx, r.db, video)
}

func (r *VideoBaseinfoRepo) SearchVideosByKeyword(ctx context.Context, keyword string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	return videobasetable.SearchVideosByKeyword(ctx, r.db, keyword, pageNum, pageSize)
}

func (r *VideoBaseinfoRepo) GetVideosByIDs(ctx context.Context, videoIDs []string) ([]types.VideoBaseinfo, error) {
	return videobasetable.GetVideosByIDs(ctx, r.db, videoIDs)
}

func (r *VideoBaseinfoRepo) GetVideosByAuthorID(ctx context.Context, authorID string, pageNum, pageSize int32) ([]types.VideoBaseinfo, int64, error) {
	return videobasetable.GetVideosByAuthorID(ctx, r.db, authorID, pageNum, pageSize)
}

type VideoPopularRepo struct {
	db *gorm.DB
}

func NewVideoPopularRepo(db *gorm.DB) *VideoPopularRepo {
	return &VideoPopularRepo{db: db}
}

func (r *VideoPopularRepo) CreatePopularVideo(ctx context.Context, videoID string) error {
	return videopopulartable.CreatePopularVideo(ctx, r.db, videoID)
}

func (r *VideoPopularRepo) IncreaseVideoVisitCount(ctx context.Context, videoID string, delta int64) error {
	return videopopulartable.IncreaseVideoVisitCount(ctx, r.db, videoID, delta)
}

func (r *VideoPopularRepo) UpdateVideoLikeCount(ctx context.Context, videoID string, delta int64) error {
	return videopopulartable.UpdateVideoLikeCount(ctx, r.db, videoID, delta)
}

func (r *VideoPopularRepo) GetPopularVideoIDsByVisitCount(ctx context.Context, pageNum, pageSize int32) ([]string, int64, error) {
	return videopopulartable.GetPopularVideoIDsByVisitCount(ctx, r.db, pageNum, pageSize)
}

type CommentRepo struct {
	db *gorm.DB
}

func NewCommentRepo(db *gorm.DB) *CommentRepo {
	return &CommentRepo{db: db}
}

func (r *CommentRepo) CreateComment(ctx context.Context, comment *types.CommentBaseinfo) error {
	return commenttable.CreateComment(ctx, r.db, comment)
}

func (r *CommentRepo) DeleteCommentByID(ctx context.Context, commentID string) error {
	return commenttable.DeleteCommentByID(ctx, r.db, commentID)
}

func (r *CommentRepo) GetCommentsByVideoID(ctx context.Context, videoID string, pageNumber, pageSize int32) ([]types.CommentBaseinfo, int64, error) {
	return commenttable.GetCommentsByVideoID(ctx, r.db, videoID, pageNumber, pageSize)
}

type VideoLikerRepo struct {
	db *gorm.DB
}

func NewVideoLikerRepo(db *gorm.DB) *VideoLikerRepo {
	return &VideoLikerRepo{db: db}
}

func (r *VideoLikerRepo) LikeVideo(ctx context.Context, userID, videoID string) error {
	return videolikertable.LikeVideo(ctx, r.db, userID, videoID)
}

func (r *VideoLikerRepo) CancelLikeVideo(ctx context.Context, userID, videoID string) error {
	return videolikertable.CancelLikeVideo(ctx, r.db, userID, videoID)
}

func (r *VideoLikerRepo) GetLikedVideoIDsByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]string, int64, error) {
	return videolikertable.GetLikedVideoIDsByUserID(ctx, r.db, userID, pageNumber, pageSize)
}

type UserFollowRepo struct {
	db *gorm.DB
}

func NewUserFollowRepo(db *gorm.DB) *UserFollowRepo {
	return &UserFollowRepo{db: db}
}

func (r *UserFollowRepo) CreateUserFollow(ctx context.Context, followerID, userID string) error {
	return userfollowtable.CreateUserFollow(ctx, r.db, followerID, userID)
}

func (r *UserFollowRepo) UpdateUserFollowStatus(ctx context.Context, followerID, userID string, status int32) error {
	return userfollowtable.UpdateUserFollowStatus(ctx, r.db, followerID, userID, status)
}

func (r *UserFollowRepo) GetFollowingISSubriber(ctx context.Context, followerID, userID string) (bool, error) {
	return userfollowtable.GetFollowingISSubriber(ctx, r.db, followerID, userID)
}

func (r *UserFollowRepo) GetFollowingByFollowerID(ctx context.Context, followerID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	return userfollowtable.GetFollowingByFollowerID(ctx, r.db, followerID, pageNumber, pageSize)
}

func (r *UserFollowRepo) GetFansByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	return userfollowtable.GetFansByUserID(ctx, r.db, userID, pageNumber, pageSize)
}

func (r *UserFollowRepo) GetFriendByUserID(ctx context.Context, userID string, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	return userfollowtable.GetFriendByUserID(ctx, r.db, userID, pageNumber, pageSize)
}
