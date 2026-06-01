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

type CommentPareantCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommentPareantCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentPareantCommentLogic {
	return &CommentPareantCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
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
