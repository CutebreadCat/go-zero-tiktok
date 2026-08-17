package video

import (
	"context"
	"time"

	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/app/gateway/api/internal/types"
	interactionpb "go_zero-tiktok/app/interaction/rpc/interaction_pb"
	videopb "go_zero-tiktok/app/video/rpc/video_pb"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type GetFeedVideoLogic struct {
	*logger.ContextLogger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFeedVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFeedVideoLogic {
	return &GetFeedVideoLogic{
		ContextLogger: logger.WithContext(ctx),
		ctx:           ctx,
		svcCtx:        svcCtx,
	}
}

func (l *GetFeedVideoLogic) GetFeedVideo(req *types.FeedVideoRequest) (resp *types.FeedVideoResponse, err error) {
	// 从登录态取 viewer_id 透传给 video.rpc，合并 feed:global + feed:inbox:{uid}
	// 拿不到身份（未登录）时传 0，仅返回全站候选池
	viewerID, _ := myutils.GetUserIDFromContext(l.ctx)

	rpcResp, err := l.svcCtx.VideoRpc.GetFeedVideo(l.ctx, &videopb.GetFeedVideoRequest{
		LastTime: req.LastTime,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		ViewerId: viewerID,
	})
	if err != nil {
		return nil, xerr.HandleDaoError(err, "GetFeedVideo.GetFeedVideo")
	}

	// 批量查询互动计数与 viewer 状态（失败降级，不影响 Feed 主链路）。
	interactionStats := l.batchGetInteractionStats(viewerID, rpcResp.Videos)

	items := make([]types.Item, 0, len(rpcResp.Videos))
	for _, v := range rpcResp.Videos {
		stat := interactionStats[v.VideoId]
		items = append(items, types.Item{
			Videos: types.VideoBaseinfo{
				VideoID:     v.VideoId,
				AuthorID:    v.AuthorId,
				VideoURL:    v.VideoUrl,
				CoverURL:    v.CoverUrl,
				Title:       v.Title,
				Description: v.Description,
				CreatedAt:   v.CreatedAt,
			},
			VideosPopular: types.VideoPopular{
				VideoID:       v.VideoId,
				VisitCount:    v.VisitCount,
				LikeCount:     stat.LikeCount,
				CommentCount:  stat.CommentCount,
				FavoriteCount: stat.FavoriteCount,
			},
			Liked:     stat.Liked,
			Favorited: stat.Favorited,
		})
	}

	return &types.FeedVideoResponse{
		Base:  types.BaseResponse{StatusCode: 0, StatusMsg: "ok"},
		Total: rpcResp.Total,
		Items: items,
	}, nil
}

// batchGetInteractionStats 批量拉取互动计数与当前用户状态，interaction 失败时返回默认值。
func (l *GetFeedVideoLogic) batchGetInteractionStats(viewerID int64, videos []*videopb.VideoInfo) map[int64]*interactionpb.VideoInteractionStat {
	result := make(map[int64]*interactionpb.VideoInteractionStat, len(videos))
	for _, v := range videos {
		result[v.VideoId] = &interactionpb.VideoInteractionStat{VideoId: v.VideoId}
	}
	if len(videos) == 0 {
		return result
	}

	videoIDs := make([]int64, 0, len(videos))
	for _, v := range videos {
		videoIDs = append(videoIDs, v.VideoId)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	rpcResp, err := l.svcCtx.InteractionRpc.BatchGetVideoInteractionStats(ctx, &interactionpb.BatchGetVideoInteractionStatsRequest{
		VideoIds: videoIDs,
		UserId:   viewerID,
	})
	if err != nil {
		l.Errorf("batch get interaction stats failed: %v", err)
		return result
	}

	for _, stat := range rpcResp.Stats {
		if stat != nil {
			result[stat.VideoId] = stat
		}
	}
	return result
}
