package interaction

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	interactionpb "go_zero-tiktok/app/interaction/rpc/interaction_pb/interaction_pb"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/Prometheus/logger"
)

type CommentPareantCommentLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommentPareantCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentPareantCommentLogic {
	return &CommentPareantCommentLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *CommentPareantCommentLogic) CommentPareantComment(req *types.CommentPareantCommentRequest) (resp *types.CommentPareantCommentResponse, err error) {
	UserId, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户未登录或登录已过期")
	}

	rpcResp, err := l.svcCtx.InteractionRpc.ReplyComment(l.ctx, &interactionpb.ReplyCommentRequest{
		UserId:          UserId,
		CommentText:     req.CommentText,
		ParentCommentId: req.ParentCommentID,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.CommentPareantCommentResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "评论发布成功",
		},
		CommentID: rpcResp.CommentId,
	}
	return resp, nil
}
