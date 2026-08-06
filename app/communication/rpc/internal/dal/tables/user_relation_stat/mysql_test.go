package user_relation_stat

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"go_zero-tiktok/testhelpers"
)

// TestGetByUserID_ParamError 表驱动：查询参数校验与未找到
func TestGetByUserID_ParamError(t *testing.T) {
	tests := []struct {
		name    string
		userID  int64
		wantErr bool
		isNoRow bool
	}{
		{"用户ID为零", 0, true, false},
		{"不存在的记录", 999, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.NewTestDB(t)
			_, err := GetByUserID(context.Background(), db, tt.userID)
			if !tt.wantErr {
				testhelpers.AssertNoErr(t, err)
				return
			}
			if tt.isNoRow {
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("expected sql.ErrNoRows, got %v", err)
				}
			} else {
				testhelpers.AssertInvalidParam(t, err)
			}
		})
	}
}

// TestGetOrCreate_Idempotent 获取或创建幂等：重复调用不产生多行
func TestGetOrCreate_Idempotent(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()

	s1, err := GetOrCreate(ctx, db, 100)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, s1.FollowerCount, int64(0))

	s2, err := GetOrCreate(ctx, db, 100)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, s2.UserID, int64(100))

	var count int64
	_ = db.Model(&UserRelationStat{}).Count(&count)
	testhelpers.AssertEqual(t, count, int64(1))
}

// TestIncreaseCounters 三计数原子增减 + 行不存在自动初始化 + 参数校验
func TestIncreaseCounters(t *testing.T) {
	// 粉丝数：+1 再 +2 = 3
	t.Run("follower +1+2=3", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		ctx := context.Background()
		testhelpers.AssertNoErr(t, IncreaseFollowerCount(ctx, db, 100, 1))
		testhelpers.AssertNoErr(t, IncreaseFollowerCount(ctx, db, 100, 2))
		stat, err := GetByUserID(ctx, db, 100)
		testhelpers.AssertNoErr(t, err)
		testhelpers.AssertEqual(t, stat.FollowerCount, int64(3))
		testhelpers.AssertEqual(t, stat.FollowingCount, int64(0))
	})

	// 关注数：+5 再 -2 = 3
	t.Run("following +5-2=3", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		ctx := context.Background()
		testhelpers.AssertNoErr(t, IncreaseFollowingCount(ctx, db, 100, 5))
		testhelpers.AssertNoErr(t, IncreaseFollowingCount(ctx, db, 100, -2))
		stat, err := GetByUserID(ctx, db, 100)
		testhelpers.AssertNoErr(t, err)
		testhelpers.AssertEqual(t, stat.FollowingCount, int64(3))
	})

	// 互关数：+1 再 +1 = 2
	t.Run("friend +1+1=2", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		ctx := context.Background()
		testhelpers.AssertNoErr(t, IncreaseFriendCount(ctx, db, 100, 1))
		testhelpers.AssertNoErr(t, IncreaseFriendCount(ctx, db, 100, 1))
		stat, err := GetByUserID(ctx, db, 100)
		testhelpers.AssertNoErr(t, err)
		testhelpers.AssertEqual(t, stat.FriendCount, int64(2))
	})

	// 参数校验：userID 为零
	t.Run("userID为零报错", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		err := IncreaseFollowerCount(context.Background(), db, 0, 1)
		testhelpers.AssertInvalidParam(t, err)
	})

	// 独立性：三种计数互不影响
	t.Run("计数相互独立", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		ctx := context.Background()
		testhelpers.AssertNoErr(t, IncreaseFollowerCount(ctx, db, 100, 1))
		testhelpers.AssertNoErr(t, IncreaseFollowingCount(ctx, db, 100, 2))
		testhelpers.AssertNoErr(t, IncreaseFriendCount(ctx, db, 100, 3))
		stat, err := GetByUserID(ctx, db, 100)
		testhelpers.AssertNoErr(t, err)
		testhelpers.AssertEqual(t, stat.FollowerCount, int64(1))
		testhelpers.AssertEqual(t, stat.FollowingCount, int64(2))
		testhelpers.AssertEqual(t, stat.FriendCount, int64(3))
	})
}
