package user

import (
	"context"

	"go_zero-tiktok/internal/shared/xerr"
)

type MfaSecretGenerator interface {
	GenerateSecret(ctx context.Context, userID string) (secret, url string, err error)
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

// GenerateQRCode 生成 MFA 二维码
func (s *MfaService) GenerateQRCode(ctx context.Context, userID string) (secret, url string, err error) {
	secret, url, err = s.mfaGen.GenerateSecret(ctx, userID)
	if err != nil {
		return "", "", xerr.HandleDaoError(err, "GetMfaqrcode.GenerateSecret")
	}
	if err := s.userRepo.UpdateUserMFAPendingSecret(ctx, userID, secret); err != nil {
		return "", "", xerr.HandleDaoError(err, "GetMfaqrcode.UpdateUserMFAPendingSecret")
	}
	return secret, url, nil
}

// BindMFA 绑定 MFA
func (s *MfaService) BindMFA(ctx context.Context, userID, secret, code string) error {
	if err := s.mfa.ValidateMfaCode(ctx, secret, code); err != nil {
		return xerr.NewInvalidParam("MFA验证失败")
	}
	if err := s.userRepo.EnableUserMFA(ctx, userID); err != nil {
		return xerr.HandleDaoError(err, "BindMfa.EnableUserMFA")
	}
	return nil
}
