package logic

import (
	"bytes"
	"context"

	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
	"go_zero-tiktok/pkg/storage/aliyun"
	myutils "go_zero-tiktok/pkg/utils"
	"go_zero-tiktok/pkg/xerr"

	logger "go_zero-tiktok/pkg/logger"
)

type PublishVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	*logger.ContextLogger
}

func NewPublishVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishVideoLogic {
	return &PublishVideoLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		ContextLogger: logger.WithContext(ctx),
	}
}

func (l *PublishVideoLogic) PublishVideo(in *video_pb.PublishVideoRequest) (*video_pb.PublishVideoResponse, error) {
	if in.UserId == 0 {
		return nil, xerr.NewInvalidParam("用户ID不能为空")
	}
	if in.Title == "" {
		return nil, xerr.NewInvalidParam("视频标题不能为空")
	}
	if len(in.VideoData) == 0 {
		return nil, xerr.NewInvalidParam("视频文件内容为空")
	}

	videoID := myutils.GenerateVideoID()
	videoObjectKey := aliyun.BuildObjectKey(aliyun.ObjectTypeVideo, in.UserId, videoID, in.Filename)

	if _, err := l.svcCtx.Storage.UploadFile(bytes.NewReader(in.VideoData), videoObjectKey); err != nil {
		return nil, xerr.HandleDaoError(err, "PublishVideo.UploadToOSS")
	}

	publishAt, err := l.svcCtx.VideoService.PublishVideo(l.ctx, videoID, in.UserId, videoObjectKey, "", in.Title, in.Description)
	if err != nil {
		return nil, xerr.HandleDaoError(err, "PublishVideo.CreateVideo")
	}

	return &video_pb.PublishVideoResponse{
		VideoId:     videoID,
		PublishedAt: publishAt.UnixMilli(),
	}, nil
}
