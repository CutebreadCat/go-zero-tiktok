package interaction

import (
	"context"
	"strings"

	"go_zero-tiktok/app/gateway/api/internal/logic/communication"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	interactionpb "go_zero-tiktok/app/interaction/rpc/interaction_pb"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type CommentVideoLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommentVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentVideoLogic {
	return &CommentVideoLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *CommentVideoLogic) CommentVideo(req *types.CommentVideoRequest) (resp *types.CommentVideoResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	if req.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}
	commentText := strings.TrimSpace(req.CommentText)
	if commentText == "" {
		return nil, xerr.NewInvalidParam("评论内容不能为空")
	}

	rpcResp, err := l.svcCtx.InteractionRpc.CommentVideo(l.ctx, &interactionpb.CommentVideoRequest{
		UserId:      userID,
		VideoId:     req.VideoId,
		CommentText: commentText,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "CommentVideo.CreateComment")
	}

	// 评论成功后触发热度分重算（Kafka 事件解耦）。
	l.svcCtx.TriggerHotScoreRecalc(l.ctx, req.VideoId)

	// 创建评论消息通知（非关键路径，失败仅记日志不影响主响应）。
	content := commentText
	if len(content) > 50 {
		content = content[:50] + "..."
	}
	if err := communication.CreateMessageForInteraction(l.ctx, l.svcCtx, userID, req.VideoId, "COMMENT", "收到评论", "有人评论了你的视频："+content); err != nil {
		l.Errorf("CommentVideo.CreateMessageForInteraction failed: %v", err)
	}

	return &types.CommentVideoResponse{
		Base:      types.BaseResponse{StatusCode: 0, StatusMsg: "评论发布成功"},
		CommentID: rpcResp.CommentId,
	}, nil
}
