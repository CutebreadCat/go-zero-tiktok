package myutils

import (
	"context"
	"fmt"

	"go_zero-tiktok/pkg/ctxkey"
	"go_zero-tiktok/pkg/xerr"

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

// GenerateUserID 生成用户ID
func GenerateUserID() int64 {
	return snowflakeNode.Generate().Int64()
}

// GenerateVideoID 生成视频ID
func GenerateVideoID() int64 {
	return snowflakeNode.Generate().Int64()
}

func GenerateCommentID() int64 {
	return snowflakeNode.Generate().Int64()
}

func GenerateRoomID() int64 {
	return snowflakeNode.Generate().Int64()
}

func GenerateMessageID() int64 {
	return snowflakeNode.Generate().Int64()
}

func GetUserIDFromContext(ctx context.Context) (int64, error) {
	if ctx == nil {
		return 0, xerr.New(500, "上下文为空")
	}

	if uid, ok := ctx.Value(ctxkey.UserID).(int64); ok && uid != 0 {
		return uid, nil
	}

	return 0, xerr.New(401, "用户未登录或登录已过期")
}
