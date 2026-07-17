package svc

import (
	"context"
	"io"
	"strings"

	userdomain "go_zero-tiktok/app/user/rpc/internal/domain"
	"go_zero-tiktok/app/user/rpc/internal/mfa"
	"go_zero-tiktok/pkg/jwt"
	"go_zero-tiktok/pkg/storage/aliyun"

	"github.com/west2-online/jwch"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type TokenAdapter struct{}

func (a *TokenAdapter) GenerateAccessToken(secret, userID string) (string, error) {
	return token.GenerateAccessToken(secret, userID)
}

func (a *TokenAdapter) GenerateRefreshToken(secret, userID string) (string, error) {
	return token.GenerateRefreshToken(secret, userID)
}

func (a *TokenAdapter) SaveRefreshToken(ctx context.Context, rdb interface{}, refreshToken, userID string) error {
	return token.SaveRefreshToken(ctx, rdb.(*redis.Redis), refreshToken, userID)
}

func (a *TokenAdapter) ParseToken(secret, tokenStr string) (interface{}, error) {
	return token.ParseToken(secret, tokenStr)
}

func (a *TokenAdapter) GetRefreshTokenUserID(ctx context.Context, rdb interface{}, refreshToken string) (string, error) {
	return token.GetRefreshTokenUserID(ctx, rdb.(*redis.Redis), refreshToken)
}

func (a *TokenAdapter) RotateRefreshToken(ctx context.Context, rdb interface{}, oldToken, newToken, userID string) error {
	return token.RotateRefreshToken(ctx, rdb.(*redis.Redis), oldToken, newToken, userID)
}

type MfaAdapter struct{}

func (a *MfaAdapter) ValidateMfaCode(ctx context.Context, secret, code string) error {
	return mfa.ValidateMfaCode(ctx, secret, code)
}

func (a *MfaAdapter) GenerateSecret(ctx context.Context, userID string) (string, string, error) {
	return mfa.GenerateSecret(ctx, userID)
}

type StorageAdapter struct{}

func (a *StorageAdapter) DeleteFile(objectKey string) error {
	return aliyun.DeleteFileFromOSS(objectKey)
}

func (a *StorageAdapter) UploadFile(reader io.Reader, objectKey string) (string, error) {
	return aliyun.UploadBytesToOSS(reader, objectKey)
}

type JwchClientAdapter struct {
	client *jwch.Student
}

func (a *JwchClientAdapter) Login() error {
	return a.client.Login()
}

func (a *JwchClientAdapter) GetIdentifierAndCookies() (string, string, error) {
	user, cookies, err := a.client.GetIdentifierAndCookies()
	if err != nil {
		return "", "", err
	}
	cookieParts := make([]string, 0, len(cookies))
	for _, c := range cookies {
		cookieParts = append(cookieParts, c.Name+"="+c.Value)
	}
	return user, strings.Join(cookieParts, "; "), nil
}

type JwchClientFactoryAdapter struct{}

func (f *JwchClientFactoryAdapter) NewClient(id, password string) userdomain.JwchClient {
	client := jwch.NewStudent()
	client.ID = id
	client.Password = password
	return &JwchClientAdapter{client: client}
}
