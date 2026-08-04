package interaction

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	interactionpb "go_zero-tiktok/app/interaction/rpc/interaction_pb"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type DeleteCommentLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *DeleteCommentLogic) DeleteComment(req *types.DeleteCommentRequest) (resp *types.DeleteCommentResponse, err error) {
	if req.CommentID == 0 {
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
