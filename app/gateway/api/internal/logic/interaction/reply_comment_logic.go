package interaction

import (
	"context"
	"strings"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	interactionpb "go_zero-tiktok/app/interaction/rpc/interaction_pb"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type ReplyCommentLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReplyCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplyCommentLogic {
	return &ReplyCommentLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *ReplyCommentLogic) ReplyComment(req *types.ReplyCommentRequest) (resp *types.ReplyCommentResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	if req.CommentId == 0 {
		return nil, xerr.NewInvalidParam("父评论ID不能为空")
	}
	commentText := strings.TrimSpace(req.CommentText)
	if commentText == "" {
		return nil, xerr.NewInvalidParam("回复内容不能为空")
	}

	rpcResp, err := l.svcCtx.InteractionRpc.ReplyComment(l.ctx, &interactionpb.ReplyCommentRequest{
		UserId:          userID,
		ParentCommentId: req.CommentId,
		CommentText:     commentText,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "ReplyComment.ReplyComment")
	}

	return &types.ReplyCommentResponse{
		Base:      types.BaseResponse{StatusCode: 0, StatusMsg: "回复成功"},
		CommentID: rpcResp.CommentId,
	}, nil
}
