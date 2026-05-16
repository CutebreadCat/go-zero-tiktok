package ai

import (
	"context"
	"log"

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
		log.Printf("Failed to login to JWCH: %v", err)
		return JwchLogin{}, err
	}
	user, cookie, err := jwchClient.GetIdentifierAndCookies()
	if err != nil {
		log.Printf("Failed to get identifier and cookies from JWCH: %v", err)
		return JwchLogin{}, err
	}
	jwchcookie := myutils.ParseCookieTostring(cookie)
	return JwchLogin{
		JwchId:     user,
		Jwchcookie: jwchcookie,
	}, nil

}
