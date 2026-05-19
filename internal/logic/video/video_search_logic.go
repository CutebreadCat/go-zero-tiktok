package video

import (
	"context"
	"fmt"
	"time"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VideoSearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVideoSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VideoSearchLogic {
	return &VideoSearchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VideoSearchLogic) VideoSearch(req *types.VideoSearchRequest) (resp *types.VideoSearchResponse, err error) {
	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		return nil, xerr.NewInvalidParam("每页数量不能超过100")
	}

	videos, _, err := l.svcCtx.Dal.Video.SearchVideosByKeyword(l.ctx, req.Keyword, req.PageNum, req.PageSize)
	if err != nil {
		return nil, err
	}

	videoResponses := l.svcCtx.Dal.Video.VideosToResponse(videos)

	resp = &types.VideoSearchResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "查询成功",
		},
		Videos: videoResponses,
	}

	if resp.Videos == nil {
		resp.Videos = []types.VideoBaseinfo{}
	}

	if req.Keyword == "" && len(resp.Videos) == 0 {
		resp.Base.StatusMsg = "暂无视频数据"
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		for _, video := range videos {
			if err := l.svcCtx.Dal.Popular.IncreaseVideoVisitCount(ctx, video.VideoID, 1); err != nil {
				fmt.Printf("increment visit count failed for video %s: %v\n", video.VideoID, err)
			}
		}
	}()

	return resp, nil
}
