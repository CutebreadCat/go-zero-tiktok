package domain

import (
	"context"

	"go_zero-tiktok/internal/types"
)

type UserFollowService struct {
	followRepo IUserFollowRepo
	userRepo   IUserRepo
}

func NewUserFollowService(followRepo IUserFollowRepo, userRepo IUserRepo) *UserFollowService {
	return &UserFollowService{
		followRepo: followRepo,
		userRepo:   userRepo,
	}
}

// GetFansList 获取粉丝列表（关系 ID 水合为用户详情）
func (s *UserFollowService) GetFansList(ctx context.Context, userID string, pageNum, pageSize int32) ([]types.UserBaseinfo, int64, error) {
	relations, total, err := s.followRepo.GetFansByUserID(ctx, userID, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}

	fansIDs := make([]string, 0, len(relations))
	for _, relation := range relations {
		fansIDs = append(fansIDs, relation.FollowerID)
	}

	if len(fansIDs) == 0 {
		return []types.UserBaseinfo{}, total, nil
	}

	fansList, err := s.userRepo.GetUsersByIDs(ctx, fansIDs)
	if err != nil {
		return nil, 0, err
	}

	return fansList, total, nil
}

// GetSubscriberList 获取关注列表（关系 ID 水合为用户详情）
func (s *UserFollowService) GetSubscriberList(ctx context.Context, userID string, pageNum, pageSize int32) ([]types.UserBaseinfo, int64, error) {
	relations, total, err := s.followRepo.GetFollowingByFollowerID(ctx, userID, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}

	subscriberIDs := make([]string, 0, len(relations))
	for _, relation := range relations {
		subscriberIDs = append(subscriberIDs, relation.UserID)
	}

	if len(subscriberIDs) == 0 {
		return []types.UserBaseinfo{}, total, nil
	}

	subscriberList, err := s.userRepo.GetUsersByIDs(ctx, subscriberIDs)
	if err != nil {
		return nil, 0, err
	}

	return subscriberList, total, nil
}

// GetFriendList 获取好友列表（关系 ID 水合为用户详情）
func (s *UserFollowService) GetFriendList(ctx context.Context, userID string, pageNum, pageSize int32) ([]types.UserBaseinfo, int64, error) {
	relations, total, err := s.followRepo.GetFriendByUserID(ctx, userID, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}

	friendIDs := make([]string, 0, len(relations))
	for _, relation := range relations {
		friendIDs = append(friendIDs, relation.FollowerID)
	}

	if len(friendIDs) == 0 {
		return []types.UserBaseinfo{}, total, nil
	}

	friendList, err := s.userRepo.GetUsersByIDs(ctx, friendIDs)
	if err != nil {
		return nil, 0, err
	}

	return friendList, total, nil
}

// FollowUser 关注用户
func (s *UserFollowService) FollowUser(ctx context.Context, followerID, userID string) error {
	return s.followRepo.FollowUser(ctx, followerID, userID)
}

// UnfollowUser 取消关注
func (s *UserFollowService) UnfollowUser(ctx context.Context, followerID, userID string) error {
	return s.followRepo.UnfollowUser(ctx, followerID, userID)
}
