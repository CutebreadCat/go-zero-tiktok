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

type CancelLikeCommentLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelLikeCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelLikeCommentLogic {
	return &CancelLikeCommentLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *CancelLikeCommentLogic) CancelLikeComment(req *types.CancelLikeCommentRequest) (resp *types.CancelLikeCommentResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	if req.CommentId == 0 {
		return nil, xerr.NewInvalidParam("评论ID不能为空")
	}

	_, err = l.svcCtx.InteractionRpc.LikeComment(l.ctx, &interactionpb.LikeCommentRequest{
		CommentId: req.CommentId,
		UserId:    userID,
		LikeType:  0, // 取消点赞
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "CancelLikeComment.LikeComment")
	}

	return &types.CancelLikeCommentResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "取消点赞成功"},
	}, nil
}
