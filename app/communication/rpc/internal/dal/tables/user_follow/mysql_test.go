package user_follow

import (
	"context"
	"testing"
	"go_zero-tiktok/testhelpers"

)

// TestFollowUser_ParamErrors 表驱动：关注参数校验
func TestFollowUser_ParamErrors(t *testing.T) {
	tests := []struct {
		name      string
		follower  int64
		followee  int64
		wantParam bool // true=参数错误；false=应成功
	}{
		{"follower为零", 0, 200, true},
		{"followee为零", 100, 0, true},
		{"自关报错", 100, 100, true},
		{"正常关注", 100, 200, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testhelpers.NewTestDB(t)
			err := FollowUser(context.Background(), db, tt.follower, tt.followee)
			if tt.wantParam {
				testhelpers.AssertInvalidParam(t, err)
			} else {
				testhelpers.AssertNoErr(t, err)
			}
		})
	}
}

// TestFollowUser_Success 关注成功落库（活跃行）
func TestFollowUser_Success(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	err := FollowUser(context.Background(), db, 100, 200)
	testhelpers.AssertNoErr(t, err)

	var count int64
	_ = db.Model(&UserFollow{}).Where("deleted_at IS NULL").Count(&count)
	testhelpers.AssertEqual(t, count, int64(1))
}

// TestFollowUser_Idempotency 重复关注应幂等报错且不产生新行
func TestFollowUser_Idempotency(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()
	testhelpers.AssertNoErr(t, FollowUser(ctx, db, 100, 200))
	err := FollowUser(ctx, db, 100, 200)
	testhelpers.AssertInvalidParam(t, err)

	var active int64
	_ = db.Model(&UserFollow{}).Where("deleted_at IS NULL").Count(&active)
	testhelpers.AssertEqual(t, active, int64(1))
	var total int64
	_ = db.Model(&UserFollow{}).Count(&total)
	testhelpers.AssertEqual(t, total, int64(1))
}

// TestUnfollowUser 表驱动：取关（成功软删恢复 / 未关注报错）
func TestUnfollowUser(t *testing.T) {
	t.Run("成功软删并可恢复关注", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		ctx := context.Background()
		testhelpers.AssertNoErr(t, FollowUser(ctx, db, 100, 200))
		testhelpers.AssertNoErr(t, UnfollowUser(ctx, db, 100, 200))

		// 软删除：活跃行变 0，物理行仍在
		var active int64
		_ = db.Model(&UserFollow{}).Where("deleted_at IS NULL").Count(&active)
		testhelpers.AssertEqual(t, active, int64(0))
		var total int64
		_ = db.Model(&UserFollow{}).Count(&total)
		testhelpers.AssertEqual(t, total, int64(1))

		// 再次关注：恢复软删行，活跃行变 1，不新增物理行
		testhelpers.AssertNoErr(t, FollowUser(ctx, db, 100, 200))
		var active2 int64
		_ = db.Model(&UserFollow{}).Where("deleted_at IS NULL").Count(&active2)
		testhelpers.AssertEqual(t, active2, int64(1))
		var total2 int64
		_ = db.Model(&UserFollow{}).Count(&total2)
		testhelpers.AssertEqual(t, total2, int64(1))
	})

	t.Run("未关注取关报错", func(t *testing.T) {
		db := testhelpers.NewTestDB(t)
		err := UnfollowUser(context.Background(), db, 100, 200)
		testhelpers.AssertInvalidParam(t, err)
	})
}

// TestGetFollowingByFollowerID_CursorPath 关注列表游标分页
func TestGetFollowingByFollowerID_CursorPath(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()
	for i := int64(1); i <= 5; i++ {
		testhelpers.AssertNoErr(t, FollowUser(ctx, db, 100, 200+i))
	}

	page1, total, err := GetFollowingByFollowerID(ctx, db, 100, 1, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, total, int64(5))
	testhelpers.AssertEqual(t, int64(len(page1)), int64(2))

	page2, _, err := GetFollowingByFollowerID(ctx, db, 100, 2, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(page2)), int64(2))

	page3, _, err := GetFollowingByFollowerID(ctx, db, 100, 3, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, int64(len(page3)), int64(1))
}

// TestGetFansByUserID_CursorPath 粉丝列表游标分页
func TestGetFansByUserID_CursorPath(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()
	for i := int64(1); i <= 3; i++ {
		testhelpers.AssertNoErr(t, FollowUser(ctx, db, 200+i, 100)) // 200+i 关注 100
	}

	fans, total, err := GetFansByUserID(ctx, db, 100, 1, 2)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, total, int64(3))
	testhelpers.AssertEqual(t, int64(len(fans)), int64(2))
}

// TestGetFriendByUserID_Bidirectional 互关判定（双向才为好友）
func TestGetFriendByUserID_Bidirectional(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	ctx := context.Background()
	testhelpers.AssertNoErr(t, FollowUser(ctx, db, 100, 200))
	testhelpers.AssertNoErr(t, FollowUser(ctx, db, 200, 100)) // 互关
	testhelpers.AssertNoErr(t, FollowUser(ctx, db, 100, 300)) // 单向

	friends, total, err := GetFriendByUserID(ctx, db, 100, 1, 10)
	testhelpers.AssertNoErr(t, err)
	testhelpers.AssertEqual(t, total, int64(1))
	testhelpers.AssertEqual(t, int64(len(friends)), int64(1))
	testhelpers.AssertEqual(t, friends[0].UserID, int64(200))
}

// TestGetActiveRelation_NotFound 查询不存在的活跃关系应报错
func TestGetActiveRelation_NotFound(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	_, err := GetActiveRelation(context.Background(), db, 100, 200)
	testhelpers.AssertErr(t, err)
}
