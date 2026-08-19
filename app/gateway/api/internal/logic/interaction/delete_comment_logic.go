package interaction

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	interactionpb "go_zero-tiktok/app/interaction/rpc/interaction_pb"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
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
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	if req.CommentId == 0 {
		return nil, xerr.NewInvalidParam("评论ID不能为空")
	}

	_, err = l.svcCtx.InteractionRpc.DeleteComment(l.ctx, &interactionpb.DeleteCommentRequest{
		CommentId: req.CommentId,
		UserId:    userID,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "DeleteComment.DeleteComment")
	}

	return &types.DeleteCommentResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "删除成功"},
	}, nil
}
