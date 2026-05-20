package mfa

import (
	"context"
	"go_zero-tiktok/internal/svc/xerr"

	"github.com/pquerna/otp/totp"
)

func GenerateSecret(ctx context.Context, userID string) (string, string, error) {

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuerName,
		AccountName: userID,
	})
	if err != nil {
		return "", "", xerr.New(400, generateSecretFailMsg)
	}
	return key.Secret(), key.URL(), nil

}

func ValidateMfaCode(ctx context.Context, secret, code string) error {
	ok := totp.Validate(code, secret)
	if !ok {
		return xerr.New(400, invalidMFACodeErrorMsg)
	}
	return nil
}
