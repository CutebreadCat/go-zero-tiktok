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

type LikeCommentLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLikeCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeCommentLogic {
	return &LikeCommentLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *LikeCommentLogic) LikeComment(req *types.LikeCommentRequest) (resp *types.LikeCommentResponse, err error) {
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
		LikeType:  1, // 点赞
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "LikeComment.LikeComment")
	}

	return &types.LikeCommentResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "点赞成功"},
	}, nil
}