package svc

import (
	"context"
	"io"
	"strconv"

	"go_zero-tiktok/app/user/rpc/internal/mfa"
	"go_zero-tiktok/pkg/jwt"
	"go_zero-tiktok/pkg/storage/aliyun"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type TokenAdapter struct{}

func (a *TokenAdapter) GenerateAccessToken(secret string, userID int64) (string, error) {
	return token.GenerateAccessToken(secret, strconv.FormatInt(userID, 10))
}

func (a *TokenAdapter) GenerateRefreshToken(secret string, userID int64) (string, error) {
	return token.GenerateRefreshToken(secret, strconv.FormatInt(userID, 10))
}

func (a *TokenAdapter) SaveRefreshToken(ctx context.Context, rdb *redis.Redis, refreshToken string, userID int64) error {
	return token.SaveRefreshToken(ctx, rdb, refreshToken, strconv.FormatInt(userID, 10))
}

func (a *TokenAdapter) ParseToken(secret, tokenStr string) (interface{}, error) {
	return token.ParseToken(secret, tokenStr)
}

func (a *TokenAdapter) GetRefreshTokenUserID(ctx context.Context, rdb *redis.Redis, refreshToken string) (int64, error) {
	userID, err := token.GetRefreshTokenUserID(ctx, rdb, refreshToken)
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (a *TokenAdapter) RotateRefreshToken(ctx context.Context, rdb *redis.Redis, oldToken, newToken string, userID int64) error {
	return token.RotateRefreshToken(ctx, rdb, oldToken, newToken, strconv.FormatInt(userID, 10))
}

type MfaAdapter struct{}

func (a *MfaAdapter) ValidateMfaCode(ctx context.Context, secret, code string) error {
	return mfa.ValidateMfaCode(ctx, secret, code)
}

func (a *MfaAdapter) GenerateSecret(ctx context.Context, userID int64) (string, string, error) {
	return mfa.GenerateSecret(ctx, strconv.FormatInt(userID, 10))
}

type StorageAdapter struct{}

func (a *StorageAdapter) DeleteFile(objectKey string) error {
	return aliyun.DeleteFile(objectKey)
}

func (a *StorageAdapter) UploadFile(reader io.Reader, objectKey string) (string, error) {
	return aliyun.UploadBytes(reader, objectKey)
}
