package interaction

import (
	"context"

	"go_zero-tiktok/app/gateway/api/internal/logic/communication"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	interactionpb "go_zero-tiktok/app/interaction/rpc/interaction_pb"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type LikeVideoLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLikeVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeVideoLogic {
	return &LikeVideoLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *LikeVideoLogic) LikeVideo(req *types.LikeVideoRequest) (resp *types.LikeVideoResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	if req.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}

	_, err = l.svcCtx.InteractionRpc.LikeVideo(l.ctx, &interactionpb.LikeVideoRequest{
		UserId:     userID,
		VideoId:    req.VideoId,
		ActionType: 1, // 点赞
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "LikeVideo.LikeVideo")
	}

	// 互动成功后触发热度分重算（Kafka 事件解耦）。
	l.svcCtx.TriggerHotScoreRecalc(l.ctx, req.VideoId)

	// 创建点赞消息通知（非关键路径，失败仅记日志不影响主响应）。
	if err := communication.CreateMessageForInteraction(l.ctx, l.svcCtx, userID, req.VideoId, "LIKE", "收到点赞", "有人点赞了你的视频"); err != nil {
		l.Errorf("LikeVideo.CreateMessageForInteraction failed: %v", err)
	}

	return &types.LikeVideoResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "点赞成功"},
	}, nil
}
