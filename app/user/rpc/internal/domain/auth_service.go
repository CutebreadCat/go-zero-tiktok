package user_service

import (
	"context"
	"errors"
	"strconv"

	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// TokenProvider token 生成与管理接口
type TokenProvider interface {
	GenerateAccessToken(secret string, userID int64) (string, error)
	GenerateRefreshToken(secret string, userID int64) (string, error)
	SaveRefreshToken(ctx context.Context, rdb *redis.Redis, refreshToken string, userID int64) error
	ParseToken(secret, tokenStr string) (interface{}, error)
	GetRefreshTokenUserID(ctx context.Context, rdb *redis.Redis, refreshToken string) (int64, error)
	RotateRefreshToken(ctx context.Context, rdb *redis.Redis, oldToken, newToken string, userID int64) error
}

// MfaProvider MFA 验证接口
type MfaProvider interface {
	ValidateMfaCode(ctx context.Context, secret, code string) error
}

type LoginResult struct {
	UserID       int64
	AccessToken  string
	RefreshToken string
}

type AuthService struct {
	userRepo IUserRepo
	token    TokenProvider
	mfa      MfaProvider
	secret   string
	rdb      *redis.Redis
}

func NewAuthService(userRepo IUserRepo, token TokenProvider, mfa MfaProvider, secret string, rdb *redis.Redis) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		token:    token,
		mfa:      mfa,
		secret:   secret,
		rdb:      rdb,
	}
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, username, password string) (int64, error) {
	// 查询已存在用户：查询成功说明用户名被占用；查询出错时需区分“用户不存在”与 DB 故障
	_, err := s.userRepo.GetUserByUsername(ctx, username)
	if err == nil {
		return 0, xerr.NewInvalidParam("用户名已存在，请更换后重试")
	}
	var codeErr *xerr.CodeError
	if !errors.As(err, &codeErr) {
		// 非业务错误（DB 故障等），直接返回，避免被误判为“用户名不存在”
		return 0, xerr.HandleDaoError(err, "Register.GetUserByUsername")
	}

	userID := myutils.GenerateUserID()
	if err := s.userRepo.CreateUserFromParams(ctx, userID, username, myutils.HashPassword(password), ""); err != nil {
		return 0, xerr.HandleDaoError(err, "Register.CreateUser")
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
		secret, err := s.userRepo.FindUserMFASecret(ctx, user.UserID)
		if err != nil {
			return nil, xerr.HandleDaoError(err, "Login.FindUserMFASecret")
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
	claimsUserID, err := strconv.ParseInt(cp.GetUserID(), 10, 64)
	if err != nil || userID != claimsUserID {
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
