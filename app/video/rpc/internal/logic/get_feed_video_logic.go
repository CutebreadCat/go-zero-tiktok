package logic

import (
	"context"
	"strings"

	feedpkg "go_zero-tiktok/app/video/rpc/internal/domain/feed"
	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type GetFeedVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewGetFeedVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFeedVideoLogic {
	return &GetFeedVideoLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *GetFeedVideoLogic) GetFeedVideo(in *video_pb.GetFeedVideoRequest) (*video_pb.GetFeedVideoResponse, error) {
	scene, cursor, limit := normalizeFeedParams(in)
	if limit <= 0 || limit > 100 {
		return nil, xerr.NewInvalidParam("每页数量必须在1-100之间")
	}

	// viewer_id 由 gateway 从登录态透传（rpc 内部拿不到鉴权信息）
	result, err := l.svcCtx.VideoService.GetFeedVideos(l.ctx, in.ViewerId, scene, cursor, limit)
	if err != nil {
		return nil, xerr.Wrap(err, "GetFeedVideo")
	}

	videoInfos := make([]*video_pb.VideoInfo, 0, len(result.Videos))
	for i, v := range result.Videos {
		info := &video_pb.VideoInfo{
			VideoId:     v.VideoID,
			AuthorId:    v.AuthorID,
			VideoUrl:    v.VideoURL,
			CoverUrl:    v.CoverURL,
			Title:       v.Title,
			Description: v.Description,
			CreatedAt:   v.CreatedAt,
		}
		if i < len(result.Populars) {
			info.VisitCount = result.Populars[i].VisitCount
		}
		videoInfos = append(videoInfos, info)
	}

	return &video_pb.GetFeedVideoResponse{
		Videos:     videoInfos,
		Total:      result.Total,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
	}, nil
}

// normalizeFeedParams 处理新旧参数兼容：未传 cursor 时，把 last_time 编码为 timeline 游标。
func normalizeFeedParams(in *video_pb.GetFeedVideoRequest) (scene, cursor string, limit int32) {
	scene = in.Scene
	if scene == "" {
		scene = "timeline"
	}

	limit = in.Limit
	if limit <= 0 {
		limit = in.PageSize
	}
	if limit <= 0 {
		limit = 20
	}

	cursor = in.Cursor
	if cursor == "" && strings.TrimSpace(in.LastTime) != "" {
		t, err := myutils.StrToTime(in.LastTime, "")
		if err == nil {
			cursor = feedpkg.EncodeTimelineCursor(&feedpkg.TimelineCursor{PublishedAt: t.UnixMilli()})
		}
	}
	return
}
