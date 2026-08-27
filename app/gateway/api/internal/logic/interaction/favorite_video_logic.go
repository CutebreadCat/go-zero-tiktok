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

type FavoriteVideoLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFavoriteVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FavoriteVideoLogic {
	return &FavoriteVideoLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *FavoriteVideoLogic) FavoriteVideo(req *types.FavoriteVideoRequest) (resp *types.FavoriteVideoResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.NewUnauthorized("用户身份信息无效，请重新登录")
	}
	if req.VideoId == 0 {
		return nil, xerr.NewInvalidParam("视频ID不能为空")
	}

	_, err = l.svcCtx.InteractionRpc.FavoriteVideo(l.ctx, &interactionpb.FavoriteVideoRequest{
		UserId:  userID,
		VideoId: req.VideoId,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "FavoriteVideo.FavoriteVideo")
	}

	// 互动成功后触发热度分重算（Kafka 事件解耦）。
	l.svcCtx.TriggerHotScoreRecalc(l.ctx, req.VideoId)

	// 创建收藏消息通知（非关键路径，失败仅记日志不影响主响应）。
	if err := communication.CreateMessageForInteraction(l.ctx, l.svcCtx, userID, req.VideoId, "LIKE", "收到收藏", "有人收藏了你的视频"); err != nil {
		l.Errorf("FavoriteVideo.CreateMessageForInteraction failed: %v", err)
	}

	return &types.FavoriteVideoResponse{
		Base: types.BaseResponse{StatusCode: 0, StatusMsg: "收藏成功"},
	}, nil
}
