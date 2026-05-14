package repository

import (
	"context"

	userbasetable "go_zero-tiktok/internal/dal/tables/user_baseinfo"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"errors"
	"go_zero-tiktok/internal/svc/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// UserToResponse 将数据库模型转换为API响应类型
func (r *UserBaseinfoRepo) UserToResponse(user *userbasetable.UserBaseinfo) types.UserBaseinfo {
	return types.UserBaseinfo{
		UserID:    user.UserID,
		Username:  user.Username,
		PhotoURL:  user.PhotoURL,
		CreatedAt: myutils.TimeToStr(user.CreatedAt, ""),
		UpdatedAt: myutils.TimeToStr(user.UpdatedAt, ""),
	}
}

// UsersToResponse 将数据库模型切片转换为API响应类型切片
func (r *UserBaseinfoRepo) UsersToResponse(users []userbasetable.UserBaseinfo) []types.UserBaseinfo {
	result := make([]types.UserBaseinfo, 0, len(users))
	for _, u := range users {
		result = append(result, r.UserToResponse(&u))
	}
	return result
}

type UserBaseinfoRepo struct {
	db *gorm.DB
}

func NewUserBaseinfoRepo(db *gorm.DB) *UserBaseinfoRepo {
	return &UserBaseinfoRepo{db: db}
}

func (r *UserBaseinfoRepo) CreateUser(ctx context.Context, user *userbasetable.UserBaseinfo) error {
	return userbasetable.CreateUser(ctx, r.db, user)
}

// CreateUserFromParams 通过参数创建用户，logic层不需要知道数据库模型
func (r *UserBaseinfoRepo) CreateUserFromParams(ctx context.Context, userID, username, password, photoURL string) error {
	user := &userbasetable.UserBaseinfo{
		UserID:   userID,
		Username: username,
		Password: password,
		PhotoURL: photoURL,
	}
	return userbasetable.CreateUser(ctx, r.db, user)
}

func (r *UserBaseinfoRepo) GetUserByID(ctx context.Context, userID string) (*userbasetable.UserBaseinfo, error) {
	return userbasetable.GetUserByID(ctx, r.db, userID)
}

func (r *UserBaseinfoRepo) GetUserByUsername(ctx context.Context, username string) (*userbasetable.UserBaseinfo, error) {
	return userbasetable.GetUserByUsername(ctx, r.db, username)
}

func (r *UserBaseinfoRepo) UpdateUserPhotoByID(ctx context.Context, userID string, photoURL string) error {
	return userbasetable.UpdateUserPhotoByID(ctx, r.db, userID, photoURL)
}

func (r *UserBaseinfoRepo) GetUsersByIDs(ctx context.Context, userIDs []string) ([]userbasetable.UserBaseinfo, error) {
	return userbasetable.GetUsersByIDs(ctx, r.db, userIDs)
}
func (r *UserBaseinfoRepo) CheckExistsMFA(ctx context.Context, userID string) (bool, error) {
	return userbasetable.CheckUserExistsMFA(ctx, r.db, userID)
}
func (r *UserBaseinfoRepo) UpdateUserMFAPendingSecret(ctx context.Context, userID string, pendingSecret string) error {
	return userbasetable.UpdateUserMFAPendingSecret(ctx, r.db, userID, pendingSecret)
}
func (r *UserBaseinfoRepo) EnableUserMFA(ctx context.Context, userID string) error {
	logger := logx.WithContext(ctx)
	if err := r.db.Transaction(
		func(tx *gorm.DB) error {
			// 1. 获取用户MFA记录
			var userMFA userbasetable.UserMFA
			err := tx.WithContext(ctx).Model(&userbasetable.UserMFA{}).Where("user_id = ?", userID).First(&userMFA).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					logger.Errorf("user mfa not found for user_id: %s", userID)
					return xerr.New(400, "用户MFA记录不存在")
				}
				logger.Errorf("enable user mfa failed: %v", err)
				return xerr.New(400, "启用用户MFA失败")
			}
			// 2. 启用用户MFA
			mfaSecret := userMFA.MFAPendingSecret
			if mfaSecret == "" {
				logger.Errorf("enable user mfa failed: pending secret is empty for user_id: %s", userID)
				return xerr.New(400, "启用用户MFA失败，待确认密钥为空")
			}
			result := tx.WithContext(ctx).Model(&userMFA).Updates(map[string]interface{}{
				"mfa_enabled":        true,
				"mfa_secret":         mfaSecret,
				"mfa_pending_secret": "",
			})
			if result.Error != nil {
				logger.Errorf("enable user mfa failed: %v", result.Error)
				return xerr.New(400, "启用用户MFA失败")
			}
			if result.RowsAffected == 0 {
				err := gorm.ErrRecordNotFound
				logger.Errorf("enable user mfa failed, user not found: %v", err)
				return xerr.New(400, "没有进行更新")
			}
			return nil
		}); err != nil {
		return err
	}

	return nil
}

func (r *UserBaseinfoRepo) FindUserMFASecret(ctx context.Context, userID string) (string, error) {
	return userbasetable.FindUserMFASecret(ctx, r.db, userID)
}
func (r *UserBaseinfoRepo) FindUserPendMFASecret(ctx context.Context, userID string) (string, error) {
	return userbasetable.FindUserPendMFASecret(ctx, r.db, userID)
}
