package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"go_zero-tiktok/internal/shared/xerr"
	myutils "go_zero-tiktok/internal/utils"

	"github.com/sashabaranov/go-openai"
	jwch "github.com/west2-online/jwch"
)

type JwchLogin struct {
	JwchId     string `json:"jwch_id"`
	Jwchcookie string `json:"jwch_cookie"`
}

// JwchLoginToolDef 定义教务处登录工具
var JwchLoginToolDef = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        jwchLoginTool,
		Description: "登录教务处系统获取用户信息和cookie，用于访问教务处相关功能。当用户需要查询成绩、课表、考试等教务处信息时调用此工具。",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				jwchIDArg: map[string]any{
					"type":        "string",
					"description": "教务处账号（学号）",
				},
				jwchPasswordArg: map[string]any{
					"type":        "string",
					"description": "教务处密码",
				},
			},
			"required": []string{jwchIDArg, jwchPasswordArg},
		},
	},
}

// HandleJwchLogin 处理教务处登录工具调用
func HandleJwchLogin(ctx context.Context, args map[string]any) (string, error) {
	jwchID, _ := args[jwchIDArg].(string)
	jwchPassword, _ := args[jwchPasswordArg].(string)

	if jwchID == "" || jwchPassword == "" {
		return "", fmt.Errorf("jwch_id and jwch_password are required")
	}

	jwchClient := jwch.NewStudent()
	jwchClient.ID = jwchID
	jwchClient.Password = jwchPassword

	err := jwchClient.Login()
	if err != nil {
		return "", xerr.Wrap(err, "HandleJwchLogin.Login")
	}

	user, cookie, err := jwchClient.GetIdentifierAndCookies()
	if err != nil {
		return "", xerr.Wrap(err, "HandleJwchLogin.GetIdentifierAndCookies")
	}

	jwchcookie := myutils.ParseCookieTostring(cookie)

	result := JwchLogin{
		JwchId:     user,
		Jwchcookie: jwchcookie,
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", xerr.Wrap(err, "HandleJwchLogin.Marshal")
	}

	return string(jsonResult), nil
}
