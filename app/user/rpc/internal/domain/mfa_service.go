package user_service

import (
	"context"

	"go_zero-tiktok/pkg/xerr"
)

type MfaSecretGenerator interface {
	GenerateSecret(ctx context.Context, userID int64) (secret, url string, err error)
}

type MfaService struct {
	userRepo IUserRepo
	mfaGen   MfaSecretGenerator
	mfa      MfaProvider
}

func NewMfaService(userRepo IUserRepo, mfaGen MfaSecretGenerator, mfa MfaProvider) *MfaService {
	return &MfaService{
		userRepo: userRepo,
		mfaGen:   mfaGen,
		mfa:      mfa,
	}
}

func (s *MfaService) GenerateQRCode(ctx context.Context, userID int64) (secret, url string, err error) {
	secret, url, err = s.mfaGen.GenerateSecret(ctx, userID)
	if err != nil {
		return "", "", xerr.HandleDaoError(err, "GetMfaqrcode.GenerateSecret")
	}
	if err := s.userRepo.UpdateUserMFAPendingSecret(ctx, userID, secret); err != nil {
		return "", "", xerr.HandleDaoError(err, "GetMfaqrcode.UpdateUserMFAPendingSecret")
	}
	return secret, url, nil
}

func (s *MfaService) BindMFA(ctx context.Context, userID int64, secret, code string) error {
	// 校验以 DB 中保存的 pending secret 为准，防止客户端传入的 secret 与生成阶段不一致
	pendingSecret, err := s.userRepo.FindUserPendMFASecret(ctx, userID)
	if err != nil {
		return xerr.HandleDaoError(err, "BindMfa.FindUserPendMFASecret")
	}
	if pendingSecret == "" {
		return xerr.NewInvalidParam("MFA 尚未生成二维码，请先获取")
	}
	if err := s.mfa.ValidateMfaCode(ctx, pendingSecret, code); err != nil {
		return xerr.NewInvalidParam("MFA验证失败")
	}
	if err := s.userRepo.EnableUserMFA(ctx, userID); err != nil {
		return xerr.HandleDaoError(err, "BindMfa.EnableUserMFA")
	}
	return nil
}
