package domain

import (
	"context"

	"go_zero-tiktok/pkg/contract"
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

// hydrateUsers 将关注关系列表中的 ID 水合为用户详情
func (s *UserFollowService) hydrateUsers(ctx context.Context, relations []types.UserFollow, pickID func(types.UserFollow) int64) ([]types.UserBaseinfo, error) {
	ids := make([]int64, 0, len(relations))
	for _, relation := range relations {
		ids = append(ids, pickID(relation))
	}
	if len(ids) == 0 {
		return []types.UserBaseinfo{}, nil
	}
	users, err := s.userRepo.GetUsersByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetFansList 获取粉丝列表（关系 ID 水合为用户详情）
func (s *UserFollowService) GetFansList(ctx context.Context, userID int64, pageNum, pageSize int32) ([]types.UserBaseinfo, int64, error) {
	relations, total, err := s.followRepo.GetFansByUserID(ctx, userID, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	fansList, err := s.hydrateUsers(ctx, relations, func(r types.UserFollow) int64 { return r.FollowerID })
	if err != nil {
		return nil, 0, err
	}
	return fansList, total, nil
}

// GetSubscriberList 获取关注列表（关系 ID 水合为用户详情）
func (s *UserFollowService) GetSubscriberList(ctx context.Context, userID int64, pageNum, pageSize int32) ([]types.UserBaseinfo, int64, error) {
	relations, total, err := s.followRepo.GetFollowingByFollowerID(ctx, userID, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	subscriberList, err := s.hydrateUsers(ctx, relations, func(r types.UserFollow) int64 { return r.UserID })
	if err != nil {
		return nil, 0, err
	}
	return subscriberList, total, nil
}

// GetFriendList 获取好友列表（关系 ID 水合为用户详情）
func (s *UserFollowService) GetFriendList(ctx context.Context, userID int64, pageNum, pageSize int32) ([]types.UserBaseinfo, int64, error) {
	relations, total, err := s.followRepo.GetFriendByUserID(ctx, userID, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	friendList, err := s.hydrateUsers(ctx, relations, func(r types.UserFollow) int64 { return r.UserID })
	if err != nil {
		return nil, 0, err
	}
	return friendList, total, nil
}

// FollowUser 关注用户
func (s *UserFollowService) FollowUser(ctx context.Context, followerID, userID int64) error {
	return s.followRepo.FollowUser(ctx, followerID, userID)
}

// UnfollowUser 取消关注
func (s *UserFollowService) UnfollowUser(ctx context.Context, followerID, userID int64) error {
	return s.followRepo.UnfollowUser(ctx, followerID, userID)
}
