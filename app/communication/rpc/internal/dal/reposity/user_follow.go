package reposity

import (
	"context"

	userfollowtable "go_zero-tiktok/app/communication/rpc/internal/dal/tables/user_follow"
	"go_zero-tiktok/pkg/contract"

	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
)

type UserFollowRepo struct {
	db *gorm.DB
}

func NewUserFollowRepo(db *gorm.DB) *UserFollowRepo {
	return &UserFollowRepo{db: db}
}

func (r *UserFollowRepo) FollowUser(ctx context.Context, followerID, userID int64) error {
	if err := userfollowtable.FollowUser(ctx, r.db, followerID, userID); err != nil {
		return pkgerrors.WithMessage(err, "UserFollowRepo.FollowUser")
	}
	return nil
}

func (r *UserFollowRepo) UnfollowUser(ctx context.Context, followerID, userID int64) error {
	if err := userfollowtable.UnfollowUser(ctx, r.db, followerID, userID); err != nil {
		return pkgerrors.WithMessage(err, "UserFollowRepo.UnfollowUser")
	}
	return nil
}

func (r *UserFollowRepo) userFollowsToResponse(relations []userfollowtable.UserFollow) []types.UserFollow {
	result := make([]types.UserFollow, 0, len(relations))
	for _, rel := range relations {
		result = append(result, types.UserFollow{
			FollowerID: rel.FollowerID,
			UserID:     rel.UserID,
		})
	}
	return result
}

func (r *UserFollowRepo) GetFollowingByFollowerID(ctx context.Context, followerID int64, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	relations, total, err := userfollowtable.GetFollowingByFollowerID(ctx, r.db, followerID, pageNumber, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "UserFollowRepo.GetFollowingByFollowerID")
	}
	return r.userFollowsToResponse(relations), total, nil
}

func (r *UserFollowRepo) GetFansByUserID(ctx context.Context, userID int64, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	relations, total, err := userfollowtable.GetFansByUserID(ctx, r.db, userID, pageNumber, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "UserFollowRepo.GetFansByUserID")
	}
	return r.userFollowsToResponse(relations), total, nil
}

func (r *UserFollowRepo) GetFriendByUserID(ctx context.Context, userID int64, pageNumber, pageSize int32) ([]types.UserFollow, int64, error) {
	relations, total, err := userfollowtable.GetFriendByUserID(ctx, r.db, userID, pageNumber, pageSize)
	if err != nil {
		return nil, 0, pkgerrors.WithMessage(err, "UserFollowRepo.GetFriendByUserID")
	}
	return r.userFollowsToResponse(relations), total, nil
}
