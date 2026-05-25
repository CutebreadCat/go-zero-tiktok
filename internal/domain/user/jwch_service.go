package user

import (
	"context"

	"go_zero-tiktok/internal/shared/xerr"
)

// JwchClient 教务处客户端接口
type JwchClient interface {
	Login() error
	GetIdentifierAndCookies() (string, string, error)
}

// JwchClientFactory 教务处客户端工厂
type JwchClientFactory interface {
	NewClient(id, password string) JwchClient
}

type JwchService struct {
	userRepo IUserRepo
	factory  JwchClientFactory
}

func NewJwchService(userRepo IUserRepo, factory JwchClientFactory) *JwchService {
	return &JwchService{
		userRepo: userRepo,
		factory:  factory,
	}
}

// Login 教务处登录并保存信息
func (s *JwchService) Login(ctx context.Context, userID, username, password string) error {
	client := s.factory.NewClient(username, password)
	if err := client.Login(); err != nil {
		return xerr.NewInvalidParam("你的账号密码无法通过教务处认证，请检查后重试")
	}
	if err := s.userRepo.UpdateUserJwchInfo(ctx, userID, username, password); err != nil {
		return xerr.HandleDaoError(err, "JwchLogin.UpdateUserJwchInfo")
	}
	return nil
}

// GetCookie 获取教务处 Cookie
func (s *JwchService) GetCookie(ctx context.Context, userID string) (identifier string, cookie string, err error) {
	jwchID, jwchPassword, err := s.userRepo.GetUserJwchInfo(ctx, userID)
	if err != nil {
		return "", "", xerr.NewInvalidParam("获取教务处信息失败，请先绑定教务处账号")
	}

	client := s.factory.NewClient(jwchID, jwchPassword)
	if err := client.Login(); err != nil {
		return "", "", xerr.NewInvalidParam("教务处登录失败，请检查账号密码是否正确")
	}

	user, rawCookie, err := client.GetIdentifierAndCookies()
	if err != nil {
		return "", "", xerr.NewInvalidParam("获取cookie失败")
	}

	return user, rawCookie, nil
}
