// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package interaction

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"

	"go_zero-tiktok/internal/svc/xerr"
	myutils "go_zero-tiktok/internal/utils"
)

type LikeCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLikeCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeCommentLogic {
	return &LikeCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LikeCommentLogic) LikeComment(req *types.LikeCommentRequest) (resp *types.LikeCommentResponse, err error) {
	userId, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		logx.Errorf("获取用户id失败: %v", err)
		return nil, xerr.New(401, "用户未登录或登录已过期")
	}
	if userId == "" {
		logx.Error("用户id为空")
		return nil, xerr.New(401, "用户未登录或登录已过期")
	}
	if req.Liketype != 1 && req.Liketype != 0 {
		logx.Errorf("无效的点赞类型: %d", req.Liketype)
		return nil, xerr.New(400, "无效的点赞类型")
	}
	if err = l.svcCtx.Dal.Comment.LikeComment(l.ctx, req.CommentID, userId, req.Liketype); err != nil {
		logx.Errorf("操作失败: %v", err)

		return nil, err
	}
	resp = &types.LikeCommentResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "操作成功",
		},
	}
	return resp, nil
}
