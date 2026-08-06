package repository

import (
	"context"
	"errors"

	userbasetable "go_zero-tiktok/app/user/rpc/internal/dal/tables/user_baseinfo"
	"go_zero-tiktok/pkg/contract"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
)

func (r *UserBaseinfoRepo) UserToResponse(user *userbasetable.UserBaseinfo) types.UserBaseinfo {
	return types.UserBaseinfo{
		UserID:    user.UserID,
		Username:  user.Username,
		PhotoURL:  user.PhotoURL,
		CreatedAt: myutils.TimeToStr(user.CreatedAt, ""),
		UpdatedAt: myutils.TimeToStr(user.UpdatedAt, ""),
	}
}

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
	if err := userbasetable.CreateUser(ctx, r.db, user); err != nil {
		return pkgerrors.WithMessage(err, "UserBaseinfoRepo.CreateUser")
	}
	return nil
}

func (r *UserBaseinfoRepo) CreateUserFromParams(ctx context.Context, userID int64, username, password, photoURL string) error {
	user := &userbasetable.UserBaseinfo{
		UserID:   userID,
		Username: username,
		Password: password,
		PhotoURL: photoURL,
		Status:   1,
	}
	if err := userbasetable.CreateUser(ctx, r.db, user); err != nil {
		return pkgerrors.WithMessage(err, "UserBaseinfoRepo.CreateUserFromParams")
	}
	return nil
}

func (r *UserBaseinfoRepo) DeleteUserByID(ctx context.Context, userID int64) error {
	if err := userbasetable.DeleteUserByID(ctx, r.db, userID); err != nil {
		return pkgerrors.WithMessage(err, "UserBaseinfoRepo.DeleteUserByID")
	}
	return nil
}

func (r *UserBaseinfoRepo) GetUserByID(ctx context.Context, userID int64) (*types.UserBaseinfo, error) {
	user, err := userbasetable.GetUserByID(ctx, r.db, userID)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "UserBaseinfoRepo.GetUserByID")
	}
	resp := r.UserToResponse(user)
	return &resp, nil
}

func (r *UserBaseinfoRepo) GetUserByUsername(ctx context.Context, username string) (*types.UserBaseinfo, error) {
	user, err := userbasetable.GetUserByUsername(ctx, r.db, username)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "UserBaseinfoRepo.GetUserByUsername")
	}
	resp := types.UserBaseinfo{
		UserID:   user.UserID,
		Username: user.Username,
		Password: user.Password,
		PhotoURL: user.PhotoURL,
	}
	return &resp, nil
}

func (r *UserBaseinfoRepo) UpdateUserPhotoByID(ctx context.Context, userID int64, photoURL string) error {
	if err := userbasetable.UpdateUserPhotoByID(ctx, r.db, userID, photoURL); err != nil {
		return pkgerrors.WithMessage(err, "UserBaseinfoRepo.UpdateUserPhotoByID")
	}
	return nil
}

func (r *UserBaseinfoRepo) GetUsersByIDs(ctx context.Context, userIDs []int64) ([]types.UserBaseinfo, error) {
	users, err := userbasetable.GetUsersByIDs(ctx, r.db, userIDs)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "UserBaseinfoRepo.GetUsersByIDs")
	}
	return r.UsersToResponse(users), nil
}

func (r *UserBaseinfoRepo) CheckExistsMFA(ctx context.Context, userID int64) (bool, error) {
	enabled, err := userbasetable.CheckUserExistsMFA(ctx, r.db, userID)
	if err != nil {
		return false, pkgerrors.WithMessage(err, "UserBaseinfoRepo.CheckExistsMFA")
	}
	return enabled, nil
}

func (r *UserBaseinfoRepo) UpdateUserMFAPendingSecret(ctx context.Context, userID int64, pendingSecret string) error {
	if err := userbasetable.UpdateUserMFAPendingSecret(ctx, r.db, userID, pendingSecret); err != nil {
		return pkgerrors.WithMessage(err, "UserBaseinfoRepo.UpdateUserMFAPendingSecret")
	}
	return nil
}

func (r *UserBaseinfoRepo) EnableUserMFA(ctx context.Context, userID int64) error {
	if err := r.db.Transaction(func(tx *gorm.DB) error {
		var user userbasetable.UserBaseinfo
		err := tx.WithContext(ctx).Model(&userbasetable.UserBaseinfo{}).Where("user_id = ?", userID).First(&user).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return xerr.NewInvalidParam("用户MFA记录不存在")
			}
			return xerr.Wrap(err, "query user mfa failed")
		}

		mfaSecret := user.MFAPendingSecret
		if mfaSecret == "" {
			return xerr.NewInvalidParam("启用用户MFA失败，待确认密钥为空")
		}

		result := tx.WithContext(ctx).Model(&user).Updates(map[string]interface{}{
			"mfa_enabled":        true,
			"mfa_secret":         mfaSecret,
			"mfa_pending_secret": "",
		})
		if result.Error != nil {
			return xerr.Wrap(result.Error, "update user mfa failed")
		}
		if result.RowsAffected == 0 {
			return xerr.NewInvalidParam("没有进行更新")
		}
		return nil
	}); err != nil {
		return pkgerrors.WithMessage(err, "UserBaseinfoRepo.EnableUserMFA")
	}
	return nil
}

func (r *UserBaseinfoRepo) FindUserMFASecret(ctx context.Context, userID int64) (string, error) {
	secret, err := userbasetable.FindUserMFASecret(ctx, r.db, userID)
	if err != nil {
		return "", pkgerrors.WithMessage(err, "UserBaseinfoRepo.FindUserMFASecret")
	}
	return secret, nil
}

func (r *UserBaseinfoRepo) FindUserPendMFASecret(ctx context.Context, userID int64) (string, error) {
	secret, err := userbasetable.FindUserPendMFASecret(ctx, r.db, userID)
	if err != nil {
		return "", pkgerrors.WithMessage(err, "UserBaseinfoRepo.FindUserPendMFASecret")
	}
	return secret, nil
}
