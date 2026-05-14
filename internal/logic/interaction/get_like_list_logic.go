// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package interaction

import (
	"context"

	"go_zero-tiktok/internal/svc"
	"go_zero-tiktok/internal/svc/xerr"
	"go_zero-tiktok/internal/types"
	myutils "go_zero-tiktok/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLikeListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLikeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLikeListLogic {
	return &GetLikeListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLikeListLogic) GetLikeList(req *types.GetLikeListRequest) (resp *types.GetLikeListResponse, err error) {
	userID, err := myutils.GetUserIDFromContext(l.ctx)
	if err != nil {
		return nil, xerr.New(401, "用户身份信息无效，请重新登录")
	}

	videoIDs, total, err := l.svcCtx.Dal.VideoLiker.GetLikedVideoIDsByUserID(l.ctx, userID, req.PageNumber, req.PageSize)
	if err != nil {
		return nil, xerr.New(1002, "获取点赞列表失败，请稍后重试")
	}

	videos, err := l.svcCtx.Dal.Video.GetVideosByIDs(l.ctx, videoIDs)
	if err != nil {
		return nil, xerr.New(1002, "获取点赞视频信息失败，请稍后重试")
	}

	// 将数据库模型转换为响应类型
	videoList := make([]types.VideoBaseinfo, 0, len(videos))
	for _, video := range videos {
		videoList = append(videoList, types.VideoBaseinfo{
			VideoID:     video.VideoID,
			AuthorID:    video.AuthorID,
			VideoURL:    video.VideoURL,
			CoverURL:    video.CoverURL,
			Title:       video.Title,
			Description: video.Description,
			CreatedAt:   myutils.TimeToStr(video.CreatedAt, ""),
			UpdatedAt:   myutils.TimeToStr(video.UpdatedAt, ""),
		})
	}

	resp = &types.GetLikeListResponse{
		Base: types.BaseResponse{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		VideoList: videoList,
		LikeCount: int32(total),
	}

	return
}
