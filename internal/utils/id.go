package myutils

import (
	"context"
	"fmt"
	"go_zero-tiktok/internal/svc/xerr"

	"github.com/bwmarrin/snowflake"
)

// snowflakeNode 全局雪花ID生成器实例
var snowflakeNode *snowflake.Node

func init() {
	// 初始化雪花节点,节点ID为1
	var err error
	snowflakeNode, err = snowflake.NewNode(1)
	if err != nil {
		panic(fmt.Sprintf("初始化雪花ID生成器失败: %v", err))
	}
}

// GenerateUserID 生成用户ID,格式: u + 雪花ID
// 例如: u123456789012345678
func GenerateUserID() string {
	return fmt.Sprintf("u%d", snowflakeNode.Generate().Int64())
}

// GenerateVideoID 生成视频ID,格式: v + 雪花ID
// 例如: v123456789012345678
func GenerateVideoID() string {
	return fmt.Sprintf("v%d", snowflakeNode.Generate().Int64())
}

func GenerateCommentID() string {
	return fmt.Sprintf("c%d", snowflakeNode.Generate().Int64())
}

func GetUserIDFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", xerr.New(500, "上下文为空")
	}

	keys := []string{"user_id", "userId", "uid", "UserID"}
	for _, key := range keys {
		if v := ctx.Value(key); v != nil {
			if uid, ok := v.(string); ok && uid != "" {
				return uid, nil
			}
		}
	}

	return "", xerr.New(401, "用户未登录或登录已过期")
}
