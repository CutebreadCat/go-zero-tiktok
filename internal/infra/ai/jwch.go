package ai

import (
	"context"
	"go_zero-tiktok/internal/svc/xerr"

	myutils "go_zero-tiktok/internal/utils"

	jwch "github.com/west2-online/jwch"
)

type JwchLogin struct {
	JwchId     string `json:"jwch_id"`
	Jwchcookie string `json:"jwch_cookie"`
}

func JwchLoginFunc(ctx context.Context, userid, password string) (JwchLogin, error) {
	jwchClient := jwch.NewStudent()
	jwchClient.ID = userid
	jwchClient.Password = password
	err := jwchClient.Login()
	if err != nil {
		return JwchLogin{}, xerr.Wrap(err, "JwchLoginFunc.Login")
	}
	user, cookie, err := jwchClient.GetIdentifierAndCookies()
	if err != nil {
		return JwchLogin{}, xerr.Wrap(err, "JwchLoginFunc.GetIdentifierAndCookies")
	}
	jwchcookie := myutils.ParseCookieTostring(cookie)
	return JwchLogin{
		JwchId:     user,
		Jwchcookie: jwchcookie,
	}, nil

}
