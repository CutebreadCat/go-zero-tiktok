package token

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func refreshTokenKey(refreshToken string) string {
	return RefreshPrefix + refreshToken
}

func SaveRefreshToken(ctx context.Context, rdb *redis.Redis, refreshToken, userID string) error {
	logger := logx.WithContext(ctx)
	if err := rdb.SetexCtx(ctx, refreshTokenKey(refreshToken), userID, int((24 * time.Hour).Seconds())); err != nil {
		logger.Errorf("save refresh token failed: %v", err)
		return err
	}

	return nil
}

func GetRefreshTokenUserID(ctx context.Context, rdb *redis.Redis, refreshToken string) (string, error) {
	logger := logx.WithContext(ctx)
	userID, err := rdb.GetCtx(ctx, refreshTokenKey(refreshToken))
	if err != nil {
		logger.Errorf("get refresh token failed: %v", err)
		return "", err
	}
	if userID == "" {
		err := errors.New("refresh token not found")
		logger.Errorf("get refresh token failed: %v", err)
		return "", err
	}

	return userID, nil
}

func DeleteRefreshToken(ctx context.Context, rdb *redis.Redis, refreshToken string) error {
	logger := logx.WithContext(ctx)
	if _, err := rdb.DelCtx(ctx, refreshTokenKey(refreshToken)); err != nil {
		logger.Errorf("delete refresh token failed: %v", err)
		return err
	}

	return nil
}

func RotateRefreshToken(ctx context.Context, rdb *redis.Redis, oldRefreshToken, newRefreshToken, userID string) error {
	if err := DeleteRefreshToken(ctx, rdb, oldRefreshToken); err != nil {
		return err
	}

	return SaveRefreshToken(ctx, rdb, newRefreshToken, userID)
}
