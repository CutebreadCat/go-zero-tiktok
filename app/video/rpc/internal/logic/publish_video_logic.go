package logic

import (
	"bytes"
	"context"

	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb/video_pb"
	"go_zero-tiktok/internal/shared/xerr"
	myutils "go_zero-tiktok/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishVideoLogic {
	return &PublishVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PublishVideoLogic) PublishVideo(in *video_pb.PublishVideoRequest) (*video_pb.PublishVideoResponse, error) {
	if in.UserId == "" {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}
	if in.Title == "" {
		return nil, xerr.NewInvalidParam("视频标题不能为空")
	}
	if len(in.VideoData) == 0 {
		return nil, xerr.NewInvalidParam("视频文件内容为空")
	}

	videoID := myutils.GenerateVideoID()
	objectKey := in.UserId + "/" + videoID + "/" + in.Filename

	videoURL, err := l.svcCtx.Storage.UploadFile(bytes.NewReader(in.VideoData), objectKey)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "PublishVideo.UploadToOSS")
	}

	if err := l.svcCtx.VideoService.PublishVideo(l.ctx, videoID, in.UserId, videoURL, "", in.Title, in.Description); err != nil {
		return nil, xerr.HandleDaoError(err, "PublishVideo.CreateVideo")
	}

	return &video_pb.PublishVideoResponse{
		VideoId: videoID,
	}, nil
}
