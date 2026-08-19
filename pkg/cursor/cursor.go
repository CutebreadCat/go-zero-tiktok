package cursor

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"go_zero-tiktok/pkg/xerr"
)

// Encode 将 payload 序列化为 JSON，并使用 URL-safe base64 编码生成游标字符串。
// payload 不能为 nil，否则返回无效游标错误。
func Encode(payload any) (string, error) {
	if payload == nil {
		return "", xerr.NewInvalidParam("cursor payload is nil")
	}

	content, err := json.Marshal(payload)
	if err != nil {
		return "", xerr.NewInvalidParam("invalid cursor payload")
	}

	return base64.RawURLEncoding.EncodeToString(content), nil
}

// Decode 解码游标字符串并将 JSON 反序列化到 payload 中。
// 支持 RawURL 和 Std 两种 base64 编码，便于兼容旧数据。
// 调用方应在空游标（首页）场景下自行处理，不调用本函数。
func Decode(raw string, payload any) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return xerr.NewInvalidParam("cursor is empty")
	}

	content, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		content, err = base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return xerr.NewInvalidParam("invalid cursor encoding")
		}
	}

	if err := json.Unmarshal(content, payload); err != nil {
		return xerr.NewInvalidParam("invalid cursor content")
	}

	return nil
}
