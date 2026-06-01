package user_service

import (
	"context"

	"go_zero-tiktok/internal/shared/xerr"
	myutils "go_zero-tiktok/internal/utils"
)

// TokenProvider token 生成与管理接口
type TokenProvider interface {
	GenerateAccessToken(secret, userID string) (string, error)
	GenerateRefreshToken(secret, userID string) (string, error)
	SaveRefreshToken(ctx context.Context, rdb interface{}, refreshToken, userID string) error
	ParseToken(secret, tokenStr string) (interface{}, error)
	GetRefreshTokenUserID(ctx context.Context, rdb interface{}, refreshToken string) (string, error)
	RotateRefreshToken(ctx context.Context, rdb interface{}, oldToken, newToken, userID string) error
}

// MfaProvider MFA 验证接口
type MfaProvider interface {
	ValidateMfaCode(ctx context.Context, secret, code string) error
}

type LoginResult struct {
	UserID       string
	AccessToken  string
	RefreshToken string
}

type AuthService struct {
	userRepo IUserRepo
	token    TokenProvider
	mfa      MfaProvider
	secret   string
	rdb      interface{}
}

func NewAuthService(userRepo IUserRepo, token TokenProvider, mfa MfaProvider, secret string, rdb interface{}) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		token:    token,
		mfa:      mfa,
		secret:   secret,
		rdb:      rdb,
	}
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, username, password string) (string, error) {
	if _, err := s.userRepo.GetUserByUsername(ctx, username); err == nil {
		return "", xerr.NewInvalidParam("用户名已存在，请更换后重试")
	}

	userID := myutils.GenerateUserID()
	if err := s.userRepo.CreateUserFromParams(ctx, userID, username, myutils.HashPassword(password), ""); err != nil {
		return "", xerr.HandleDaoError(err, "Register.CreateUser")
	}

	return userID, nil
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, username, password, mfaCode string) (*LoginResult, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, xerr.NewInvalidParam("用户名或密码错误")
	}

	if !myutils.CompareHashAndPassword(password, user.Password) {
		return nil, xerr.NewInvalidParam("用户名或密码错误")
	}

	mfaOk, err := s.userRepo.CheckExistsMFA(ctx, user.UserID)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "Login.CheckExistsMFA")
	}
	if mfaOk {
		if mfaCode == "" {
			return nil, xerr.NewInvalidParam("MFA 代码不能为空")
		}
		secret, err := s.userRepo.FindUserPendMFASecret(ctx, user.UserID)
		if err != nil {
			return nil, xerr.HandleDaoError(err, "Login.FindUserPendMFASecret")
		}
		if err := s.mfa.ValidateMfaCode(ctx, secret, mfaCode); err != nil {
			return nil, xerr.NewInvalidParam("MFA 验证失败")
		}
	}

	accessToken, err := s.token.GenerateAccessToken(s.secret, user.UserID)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "Login.GenerateAccessToken")
	}

	refreshToken, err := s.token.GenerateRefreshToken(s.secret, user.UserID)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "Login.GenerateRefreshToken")
	}

	if err := s.token.SaveRefreshToken(ctx, s.rdb, refreshToken, user.UserID); err != nil {
		return nil, xerr.HandleDaoError(err, "Login.SaveRefreshToken")
	}

	return &LoginResult{
		UserID:       user.UserID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(ctx context.Context, oldRefreshToken string) (accessToken, refreshToken string, err error) {
	claims, err := s.token.ParseToken(s.secret, oldRefreshToken)
	if err != nil {
		return "", "", xerr.NewUnauthorized("解析刷新令牌失败")
	}
	// 类型断言获取 claims
	type claimParser interface {
		GetTokenType() string
		GetUserID() string
	}
	cp, ok := claims.(claimParser)
	if !ok {
		return "", "", xerr.NewUnauthorized("无效的刷新令牌")
	}
	if cp.GetTokenType() != "refresh" {
		return "", "", xerr.NewUnauthorized("无效的刷新令牌")
	}

	userID, err := s.token.GetRefreshTokenUserID(ctx, s.rdb, oldRefreshToken)
	if err != nil {
		return "", "", xerr.NewUnauthorized("获取用户ID失败")
	}
	if userID != cp.GetUserID() {
		return "", "", xerr.NewUnauthorized("刷新令牌不匹配")
	}

	newAccessToken, err := s.token.GenerateAccessToken(s.secret, userID)
	if err != nil {
		return "", "", xerr.HandleDaoError(err, "RefreshToken.GenerateAccessToken")
	}

	newRefreshToken, err := s.token.GenerateRefreshToken(s.secret, userID)
	if err != nil {
		return "", "", xerr.HandleDaoError(err, "RefreshToken.GenerateRefreshToken")
	}

	if err := s.token.RotateRefreshToken(ctx, s.rdb, oldRefreshToken, newRefreshToken, userID); err != nil {
		return "", "", xerr.HandleDaoError(err, "RefreshToken.RotateRefreshToken")
	}

	return newAccessToken, newRefreshToken, nil
}
