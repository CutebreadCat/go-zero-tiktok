package user

import (
	"context"
	"errors"
	"testing"

	"go_zero-tiktok/internal/config"
	userdomain "go_zero-tiktok/internal/domain/user"
	"go_zero-tiktok/internal/middleware/token"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
)

func TestRefreshTokenLogic_RefreshToken_Table(t *testing.T) {
	secret := "test-secret"
	svcCtx := &svc.ServiceContext{
		Config: config.Config{
			Auth: config.AuthConfig{AccessSecret: secret},
		},
		Rdb:             nil,
		UserAuthService: userdomain.NewAuthService(nil, &svc.TokenAdapter{}, nil, secret, nil),
	}
	logic := NewRefreshTokenLogic(context.Background(), svcCtx)

	accessToken, err := token.GenerateAccessToken(secret, "user-1")
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}

	cases := []struct {
		name       string
		req        *types.RefreshTokenRequest
		expectCode int
	}{
		{
			name:       "empty_refresh_token",
			req:        &types.RefreshTokenRequest{RefreshToken: ""},
			expectCode: xerr.Unauthorized,
		},
		{
			name:       "invalid_token",
			req:        &types.RefreshTokenRequest{RefreshToken: "not-a-token"},
			expectCode: xerr.Unauthorized,
		},
		{
			name:       "access_token_used_as_refresh",
			req:        &types.RefreshTokenRequest{RefreshToken: accessToken},
			expectCode: xerr.Unauthorized,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := logic.RefreshToken(tc.req)
			if err == nil {
				t.Fatalf("expected error, got response: %+v", resp)
			}
			var codeErr *xerr.CodeError
			if !errors.As(err, &codeErr) {
				t.Fatalf("expected CodeError, got %T: %v", err, err)
			}
			if codeErr.Code != tc.expectCode {
				t.Fatalf("expected code %d, got %d", tc.expectCode, codeErr.Code)
			}
		})
	}
}
