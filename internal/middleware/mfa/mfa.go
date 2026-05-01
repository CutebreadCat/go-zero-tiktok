package mfa

import (
	"context"
	"go_zero-tiktok/internal/svc/xerr"

	"github.com/pquerna/otp/totp"
)

func GenerateSecret(ctx context.Context, userID string) (string, string, error) {

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "GoZeroTiktok",
		AccountName: userID,
	})
	if err != nil {
		return "", "", xerr.New(400, "生成MFA密钥失败")
	}
	return key.Secret(), key.URL(), nil

}

func ValidateMfaCode(ctx context.Context, secret, code string) error {
	ok := totp.Validate(code, secret)
	if !ok {
		return xerr.New(400, "MFA验证码无效")
	}
	return nil
}
