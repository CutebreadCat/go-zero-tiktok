package interaction

import (
	"context"

	interactionpb "go_zero-tiktok/app/interaction/rpc/interaction_pb/interaction_pb"
	"go_zero-tiktok/internal/shared/xerr"
	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCommentLogic) DeleteComment(req *types.DeleteCommentRequest) (resp *types.DeleteCommentResponse, err error) {
	if req.CommentID == "" {
		return nil, xerr.NewInvalidParam("评论ID不能为空")
	}
	userid, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("获取用户ID失败")
	}

	_, err = l.svcCtx.InteractionRpc.DeleteComment(l.ctx, &interactionpb.DeleteCommentRequest{
		CommentId: req.CommentID,
		UserId:    userid,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.DeleteCommentResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "删除评论成功"},
	}

	return resp, nil
}
