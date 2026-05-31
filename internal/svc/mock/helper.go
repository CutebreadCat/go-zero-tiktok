package mock

import (
	"context"
	"errors"
	"io"
	"strings"

	"go_zero-tiktok/internal/config"
	chatdomain "go_zero-tiktok/internal/domain/chat"
	commentdomain "go_zero-tiktok/internal/domain/comment"
	userdomain "go_zero-tiktok/internal/domain/user"
	videodomain "go_zero-tiktok/internal/domain/video"
	"go_zero-tiktok/internal/svc"
)

func NewServiceContext(
	userRepo *UserRepo,
	videoRepo *VideoRepo,
	popularRepo *PopularRepo,
	commentRepo *CommentRepo,
	videoLikerRepo *VideoLikerRepo,
	userFollowRepo *UserFollowRepo,
	chatRepo *ChatRepo,
) *svc.ServiceContext {
	repos := &svc.Repositories{
		User:       userRepo,
		Video:      videoRepo,
		Popular:    popularRepo,
		Comment:    commentRepo,
		VideoLiker: videoLikerRepo,
		UserFollow: userFollowRepo,
		Chat:       chatRepo,
	}

	return &svc.ServiceContext{
		Config: config.Config{
			Auth: config.AuthConfig{
				AccessSecret: "test-secret",
			},
		},
		Dal:                repos,
		VideoService:       videodomain.NewVideoService(videoRepo, popularRepo, videoLikerRepo),
		CommentService:     commentdomain.NewCommentService(commentRepo, popularRepo),
		UserFollowService:  userdomain.NewUserFollowService(userFollowRepo, userRepo),
		ChatService:        chatdomain.NewChatService(chatRepo),
		UserAuthService:    userdomain.NewAuthService(userRepo, fakeTokenProvider{}, fakeMfaProvider{}, "test-secret", nil),
		UserMfaService:     userdomain.NewMfaService(userRepo, fakeMfaProvider{}, fakeMfaProvider{}),
		UserProfileService: userdomain.NewProfileService(userRepo, fakeObjectStorage{}),
		UserJwchService:    userdomain.NewJwchService(userRepo, fakeJwchFactory{}),
	}
}

type fakeClaims struct {
	tokenType string
	userID    string
}

func (c fakeClaims) GetTokenType() string {
	return c.tokenType
}

func (c fakeClaims) GetUserID() string {
	return c.userID
}

type fakeTokenProvider struct{}

func (fakeTokenProvider) GenerateAccessToken(secret, userID string) (string, error) {
	return "access-" + userID, nil
}

func (fakeTokenProvider) GenerateRefreshToken(secret, userID string) (string, error) {
	return "refresh-" + userID, nil
}

func (fakeTokenProvider) SaveRefreshToken(ctx context.Context, rdb interface{}, refreshToken, userID string) error {
	return nil
}

func (fakeTokenProvider) ParseToken(secret, tokenStr string) (interface{}, error) {
	switch {
	case strings.HasPrefix(tokenStr, "refresh-"):
		return fakeClaims{tokenType: "refresh", userID: strings.TrimPrefix(tokenStr, "refresh-")}, nil
	case strings.HasPrefix(tokenStr, "access-"):
		return fakeClaims{tokenType: "access", userID: strings.TrimPrefix(tokenStr, "access-")}, nil
	default:
		return nil, errors.New("invalid token")
	}
}

func (fakeTokenProvider) GetRefreshTokenUserID(ctx context.Context, rdb interface{}, refreshToken string) (string, error) {
	if !strings.HasPrefix(refreshToken, "refresh-") {
		return "", errors.New("invalid refresh token")
	}
	return strings.TrimPrefix(refreshToken, "refresh-"), nil
}

func (fakeTokenProvider) RotateRefreshToken(ctx context.Context, rdb interface{}, oldToken, newToken, userID string) error {
	return nil
}

type fakeMfaProvider struct{}

func (fakeMfaProvider) ValidateMfaCode(ctx context.Context, secret, code string) error {
	if code == "invalid" {
		return errors.New("invalid mfa code")
	}
	return nil
}

func (fakeMfaProvider) GenerateSecret(ctx context.Context, userID string) (string, string, error) {
	return "test-secret", "otpauth://test/" + userID, nil
}

type fakeObjectStorage struct{}

func (fakeObjectStorage) DeleteFile(objectKey string) error {
	return nil
}

func (fakeObjectStorage) UploadFile(reader io.Reader, objectKey string) (string, error) {
	return "https://example.com/" + objectKey, nil
}

type fakeJwchClient struct{}

func (fakeJwchClient) Login() error {
	return nil
}

func (fakeJwchClient) GetIdentifierAndCookies() (string, string, error) {
	return "test-user", "SESSION=test", nil
}

type fakeJwchFactory struct{}

func (fakeJwchFactory) NewClient(id, password string) userdomain.JwchClient {
	return fakeJwchClient{}
}
